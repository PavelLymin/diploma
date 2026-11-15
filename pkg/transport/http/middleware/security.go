package middleware

import "github.com/gin-gonic/gin"

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Cache-Control", "no-store")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}
