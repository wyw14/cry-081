package httptransport

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wyw14/cry-081/internal/middleware"
)

type Readiness interface {
	Ready() bool
}

type endpoint struct {
	verb    string
	pattern string
	handle  gin.HandlerFunc
}

func (e endpoint) mount(group *gin.RouterGroup) {
	group.Handle(e.verb, e.pattern, e.handle)
}

type routeCatalog struct {
	handlers  *Handlers
	users     middleware.UserReader
	readiness Readiness
}

func NewRouter(handlers *Handlers, users middleware.UserReader, readiness Readiness, allowedOrigin string, rateLimit int, requestTimeout time.Duration) *gin.Engine {
	engine := gin.New()
	requestPolicy := middleware.NewRequestPolicy(allowedOrigin, rateLimit)
	engine.Use(
		gin.Recovery(),
		requestPolicy.Handle(),
		requestDeadline(requestTimeout),
	)
	catalog := routeCatalog{handlers: handlers, users: users, readiness: readiness}
	catalog.mountProbes(engine)
	catalog.mountEditorialAPI(engine.Group("/api/v1"))
	return engine
}

func (r routeCatalog) mountProbes(engine *gin.Engine) {
	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "editorial-collaboration"})
	})
	engine.GET("/readyz", func(c *gin.Context) {
		if r.readiness.Ready() {
			c.JSON(http.StatusOK, gin.H{"status": "ready"})
			return
		}
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":     "not_ready",
			"request_id": middleware.RequestIDFrom(c),
		})
	})
}

func (r routeCatalog) mountEditorialAPI(api *gin.RouterGroup) {
	public := []endpoint{
		{http.MethodPost, "/auth/register", r.handlers.Register},
		{http.MethodPost, "/auth/login", r.handlers.Login},
		{http.MethodGet, "/publications", r.handlers.SearchPublications},
	}
	for _, item := range public {
		item.mount(api)
	}

	workspace := api.Group("")
	workspace.Use(middleware.Authentication(r.users))
	protected := []endpoint{
		{http.MethodPost, "/manuscripts", r.handlers.CreateManuscript},
		{http.MethodGet, "/manuscripts", r.handlers.ListManuscripts},
		{http.MethodPost, "/manuscripts/:id/submissions", r.handlers.SubmitManuscript},
		{http.MethodPost, "/manuscripts/:id/assignments", r.handlers.Assign},
		{http.MethodPost, "/assignments/:id/reports", r.handlers.SubmitReport},
	}
	for _, item := range protected {
		item.mount(workspace)
	}
}

func requestDeadline(limit time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestContext, cancel := context.WithTimeout(c.Request.Context(), limit)
		defer cancel()
		c.Request = c.Request.WithContext(requestContext)
		c.Next()
	}
}
