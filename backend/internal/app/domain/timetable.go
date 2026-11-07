package domain

import "time"

// TimetableEvent is a synchronized entry from an external timetable
// (e.g. tidsapparat). It is read-only from crewapparat's perspective.
type TimetableEvent struct {
	ID          int64     `json:"id"`
	Source      string    `json:"source"`
	ExternalID  string    `json:"external_id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	ExternalURL string    `json:"external_url,omitempty"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time,omitempty"`
	IsHidden    bool      `json:"hidden,omitempty"`
	Category    *Category `json:"category,omitempty"`
	Location    *Location `json:"location,omitempty"`
}

type Category struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type Location struct {
	Name    string `json:"name"`
	Address string `json:"address,omitempty"`
}
