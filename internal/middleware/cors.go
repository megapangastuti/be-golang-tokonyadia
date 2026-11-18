package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	allowedMethods = "GET, POST, PUT, DELETE, OPTIONS"
	allowedHeaders = "Origin, Content-Type, Accept, Authorization, ngrok-skip-browser-warning"
)

// CORSMiddleware permits browser clients from the supplied origins to call the API.
// An empty list permits every origin. This is suitable for local sharing through an
// ngrok development tunnel because this API uses Bearer tokens, not cookies.
func CORSMiddleware(origins []string) gin.HandlerFunc {
	allowedOrigins := make(map[string]struct{}, len(origins))
	allowAnyOrigin := len(origins) == 0

	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}
		if origin == "*" {
			allowAnyOrigin = true
			continue
		}
		allowedOrigins[origin] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}

		_, originAllowed := allowedOrigins[origin]
		if !allowAnyOrigin && !originAllowed {
			if c.Request.Method == http.MethodOptions {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.Next()
			return
		}

		if allowAnyOrigin {
			c.Header("Access-Control-Allow-Origin", "*")
		} else {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Methods", allowedMethods)
		c.Header("Access-Control-Allow-Headers", allowedHeaders)
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// OriginsFromEnv converts a comma-separated CORS_ALLOWED_ORIGINS value into
// origins accepted by CORSMiddleware. An unset value allows every origin.
func OriginsFromEnv(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.Split(value, ",")
}
