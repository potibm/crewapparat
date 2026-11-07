package timetableconsumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/potibm/crewapparat/internal/app/config"
	"github.com/potibm/crewapparat/internal/app/domain"
	"github.com/potibm/crewapparat/internal/app/repository"
	"github.com/redis/go-redis/v9"
)

const (
	sourceName   = "protokolapparat-timetable"
	pelStartID   = "0"
	newMsgID     = ">"
	batchSize    = 50
	pollTimeout  = 2 * time.Second
	backoffDelay = time.Second
)

type scheduleEntry struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ExternalURL string `json:"external_url"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	IsHidden    bool   `json:"is_hidden"`
	Category    *struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	} `json:"category,omitempty"`
	Location *struct {
		Name    string `json:"name"`
		Address string `json:"address,omitempty"`
	} `json:"location,omitempty"`
}

type redisEvent struct {
	Version int             `json:"v"`
	Action  string          `json:"action"`
	Payload []scheduleEntry `json:"payload"`
}

func Run(
	ctx context.Context,
	cfg config.TimetableConfig,
	redisURL string,
	repo repository.TimetableRepository,
	logger *slog.Logger,
) error {
	if !cfg.Enabled || redisURL == "" {
		return nil
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return fmt.Errorf("invalid redis url: %w", err)
	}

	rdb := redis.NewClient(opts)
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	stream := cfg.StreamName
	if stream == "" {
		stream = "party:timetable:events"
	}

	group := cfg.ConsumerGroup
	if group == "" {
		group = "crewapparat_timetable"
	}

	consumer := cfg.ConsumerName
	if consumer == "" {
		consumer = "instance_1"
	}

	if err := rdb.XGroupCreateMkStream(ctx, stream, group, pelStartID).
		Err(); err != nil &&
		!strings.Contains(err.Error(), "BUSYGROUP") {
		return fmt.Errorf("create consumer group: %w", err)
	}

	logger.Info("timetable sync started", "stream", stream, "group", group)

	startID := pelStartID
	for ctx.Err() == nil {
		startID = runIteration(ctx, rdb, stream, group, consumer, startID, repo, logger)
	}

	logger.Info("timetable sync stopped")

	return nil
}

func runIteration(
	ctx context.Context,
	rdb *redis.Client,
	stream, group, consumer string,
	startID string,
	repo repository.TimetableRepository,
	logger *slog.Logger,
) string {
	streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  []string{stream, startID},
		Count:    batchSize,
		Block:    pollTimeout,
	}).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) || errors.Is(err, context.Canceled) {
			return startID
		}

		logger.Warn("redis read error", "error", err)
		time.Sleep(backoffDelay)

		return startID
	}

	if len(streams) == 0 || len(streams[0].Messages) == 0 {
		if startID == pelStartID {
			return newMsgID
		}

		return startID
	}

	hadFailure := processBatch(ctx, rdb, stream, group, streams, repo, logger)
	if hadFailure {
		time.Sleep(backoffDelay)

		return pelStartID
	}

	if startID == pelStartID {
		return newMsgID
	}

	return startID
}

func processBatch(
	ctx context.Context,
	rdb *redis.Client,
	stream, group string,
	streams []redis.XStream,
	repo repository.TimetableRepository,
	logger *slog.Logger,
) bool {
	hadFailure := false

	for _, st := range streams {
		for _, msg := range st.Messages {
			ok := processMessage(ctx, msg, repo, logger)
			if !ok {
				hadFailure = true

				continue
			}

			if err := rdb.XAck(ctx, stream, group, msg.ID).Err(); err != nil {
				logger.Warn("failed to ack message", "id", msg.ID, "error", err)

				hadFailure = true
			}
		}
	}

	return hadFailure
}

func processMessage(
	ctx context.Context,
	msg redis.XMessage,
	repo repository.TimetableRepository,
	logger *slog.Logger,
) bool {
	raw, ok := msg.Values["data"].(string)
	if !ok {
		logger.Warn("missing data field", "id", msg.ID)

		return true
	}

	var event redisEvent
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		logger.Warn("invalid message", "id", msg.ID, "error", err)

		return true
	}

	switch event.Action {
	case "create", "update":
		for _, p := range event.Payload {
			ev, err := mapEntry(p)
			if err != nil {
				logger.Warn("invalid entry", "error", err)

				continue
			}

			if err := repo.Save(ctx, &ev); err != nil {
				logger.Warn("save failed", "error", err)

				return false
			}
		}
	case "sync":
		events := make([]domain.TimetableEvent, 0, len(event.Payload))
		for _, p := range event.Payload {
			ev, err := mapEntry(p)
			if err != nil {
				logger.Warn("invalid entry", "error", err)

				continue
			}

			events = append(events, ev)
		}

		if _, _, _, err := repo.Sync(ctx, sourceName, events); err != nil {
			logger.Warn("sync failed", "error", err)

			return false
		}
	case "delete":
		for _, p := range event.Payload {
			if err := repo.Delete(ctx, sourceName, fmt.Sprintf("%d", p.ID)); err != nil {
				logger.Warn("delete failed", "error", err)
			}
		}
	default:
		logger.Warn("unknown action", "action", event.Action)
	}

	return true
}

func mapEntry(p scheduleEntry) (domain.TimetableEvent, error) {
	start, err := time.Parse(time.RFC3339, p.StartTime)
	if err != nil {
		return domain.TimetableEvent{}, fmt.Errorf("invalid start_time: %w", err)
	}

	end := time.Time{}

	if p.EndTime != "" {
		parsed, err := time.Parse(time.RFC3339, p.EndTime)
		if err != nil {
			return domain.TimetableEvent{}, fmt.Errorf("invalid end_time: %w", err)
		}

		end = parsed
	}

	ev := domain.TimetableEvent{
		Source:      sourceName,
		ExternalID:  fmt.Sprintf("%d", p.ID),
		Title:       p.Title,
		Description: p.Description,
		ExternalURL: p.ExternalURL,
		StartTime:   start,
		EndTime:     end,
		IsHidden:    p.IsHidden,
	}
	if p.Category != nil {
		ev.Category = &domain.Category{Name: p.Category.Name, Color: p.Category.Color}
	}

	if p.Location != nil {
		ev.Location = &domain.Location{Name: p.Location.Name, Address: p.Location.Address}
	}

	return ev, nil
}
