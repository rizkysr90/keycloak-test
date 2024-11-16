package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates the session before allowing access to protected pages
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for session_id cookie
		sessionID, err := c.Cookie("session_id")
		if err != nil || sessionID == "" {
			// No valid session found, redirect to login page
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Check for user_email cookie
		userEmail, err := c.Cookie("user_email")
		if err != nil || userEmail == "" {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Store validated user info in context for the handler to use
		c.Set("user_id", sessionID)
		c.Set("user_email", userEmail)

		c.Next()
	}
}
