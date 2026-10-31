package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/potibm/crewapparat/internal/app/config"
	"github.com/potibm/crewapparat/internal/app/domain"
)

func HeaderMiddleware(cfg config.AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		sub := DecodeDexSub(c.GetHeader(cfg.HeaderUserSub))
		if sub != "" {
			c.Set("userID", sub)

			ctxWithUser := context.WithValue(c.Request.Context(), domain.UserIDKey, sub)
			c.Request = c.Request.WithContext(ctxWithUser)
		}

		groups := c.GetHeader(cfg.HeaderUserGroups)
		if groups != "" {
			c.Set("userGroups", strings.Split(groups, ","))
		}

		c.Next()
	}
}
