package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const RequestIDKey = "request_id"

type RequestPolicy struct {
	allowedOrigin string
	requestIDs    io.Reader
	clock         func() time.Time
	budget        *minuteBudget
}

func NewRequestPolicy(allowedOrigin string, requestsPerMinute int) *RequestPolicy {
	return &RequestPolicy{
		allowedOrigin: strings.TrimSuffix(allowedOrigin, "/"),
		requestIDs:    rand.Reader,
		clock:         time.Now,
		budget:        newMinuteBudget(requestsPerMinute),
	}
}

func (p *RequestPolicy) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := p.correlationID(c.GetHeader("X-Request-ID"))
		c.Set(RequestIDKey, requestID)
		c.Header("X-Request-ID", requestID)
		p.applyBrowserPolicy(c)
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		if !p.budget.take(c.ClientIP(), p.clock().UTC()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code": "RATE_LIMITED", "message": "too many requests", "request_id": requestID,
			})
			return
		}
		c.Next()
	}
}

func (p *RequestPolicy) correlationID(candidate string) string {
	if candidate != "" && len(candidate) <= 128 {
		return candidate
	}
	var entropy [12]byte
	if _, err := io.ReadFull(p.requestIDs, entropy[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(entropy[:])
}

func (p *RequestPolicy) applyBrowserPolicy(c *gin.Context) {
	for header, value := range map[string]string{
		"X-Content-Type-Options":  "nosniff",
		"X-Frame-Options":         "DENY",
		"Referrer-Policy":         "no-referrer",
		"Content-Security-Policy": "default-src 'self'",
	} {
		c.Header(header, value)
	}
	origin := strings.TrimSuffix(c.GetHeader("Origin"), "/")
	if origin == "" || origin != p.allowedOrigin {
		return
	}
	c.Header("Access-Control-Allow-Origin", origin)
	c.Header("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type,Idempotency-Key,X-Request-ID,X-Actor-ID")
	c.Header("Vary", "Origin")
}

type minuteBudget struct {
	lock    sync.Mutex
	maximum int
	clients map[string]clientWindow
}

type clientWindow struct {
	openedAt time.Time
	used     int
}

func newMinuteBudget(maximum int) *minuteBudget {
	if maximum < 1 {
		maximum = 1
	}
	return &minuteBudget{maximum: maximum, clients: make(map[string]clientWindow)}
}

func (b *minuteBudget) take(client string, now time.Time) bool {
	b.lock.Lock()
	defer b.lock.Unlock()
	window, exists := b.clients[client]
	if !exists || now.Sub(window.openedAt) >= time.Minute {
		window = clientWindow{openedAt: now}
	}
	window.used++
	b.clients[client] = window
	return window.used <= b.maximum
}

func RequestIDFrom(c *gin.Context) string {
	requestID, _ := c.Get(RequestIDKey)
	value, _ := requestID.(string)
	return value
}
