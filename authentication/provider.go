package authentication

import "github.com/gin-gonic/gin"

// Provider is a simple authentication provider interface to integrate with Gin
type Provider interface {
	NewMiddleware() gin.HandlerFunc
}
