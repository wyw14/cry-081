package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wyw14/cry-081/internal/domain/identity"
)

const ActorKey = "actor"

type UserReader interface {
	Get(context.Context, string) (identity.User, error)
}

func Authentication(users UserReader) gin.HandlerFunc {
	return func(c *gin.Context) {
		actorID := c.GetHeader("X-Actor-ID")
		if actorID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "AUTHENTICATION_REQUIRED", "message": "authentication is required", "request_id": RequestIDFrom(c)})
			return
		}
		actor, err := users.Get(c.Request.Context(), actorID)
		if err != nil || !actor.Active {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "ACTOR_INVALID", "message": "actor is not active", "request_id": RequestIDFrom(c)})
			return
		}
		c.Set(ActorKey, actor)
		c.Next()
	}
}

func Actor(c *gin.Context) (identity.User, bool) {
	value, ok := c.Get(ActorKey)
	if !ok {
		return identity.User{}, false
	}
	actor, ok := value.(identity.User)
	return actor, ok
}
