package middleware

import (
	"net/http"
	"strings"

	"tokonyadia/internal/models"
	"tokonyadia/internal/utils"

	"github.com/gin-gonic/gin"
)

const (
	contextCustomerID = "customerID"
	contextEmail      = "email"
	contextRole       = "role"

	roleAdmin = "admin"

	errorUnauthorized = "unauthorized"
	errorForbidden    = "forbidden"
)

// AuthMiddleware validates the Bearer token and sets user claims
// into the Gin context.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := extractBearerToken(c.GetHeader("Authorization"))
		if !ok {
			abortWithError(
				c,
				http.StatusUnauthorized,
				errorUnauthorized,
				"Invalid authorization format",
			)
			return
		}

		claims, err := utils.ValidateToken(token)
		if err != nil {
			abortWithError(
				c,
				http.StatusUnauthorized,
				errorUnauthorized,
				"Invalid or expired token",
			)
			return
		}

		// Store authenticated user information in Gin context.
		c.Set(contextCustomerID, claims.CustomerID)
		c.Set(contextEmail, claims.Email)
		c.Set(contextRole, claims.Role)

		c.Next()
	}
}

// AdminMiddleware ensures that the authenticated user has admin privileges.
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString(contextRole)

		if role != roleAdmin {
			abortWithError(
				c,
				http.StatusForbidden,
				errorForbidden,
				"Admin access required",
			)
			return
		}

		c.Next()
	}
}

// extractBearerToken extracts the token from an Authorization header.
func extractBearerToken(authHeader string) (string, bool) {
	parts := strings.Fields(authHeader)

	if len(parts) != 2 {
		return "", false
	}

	if !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}

	return parts[1], true
}

// abortWithError sends a standardized error response and aborts the request.
func abortWithError(
	c *gin.Context,
	statusCode int,
	errorCode string,
	message string,
) {
	c.AbortWithStatusJSON(statusCode, models.ErrorResponse{
		Error:   errorCode,
		Message: message,
	})
}
