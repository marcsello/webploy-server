package authorization

import "github.com/gin-gonic/gin"

// Provider is an Authorization provider
type Provider interface {
	NewMiddleware(acts ...string) gin.HandlerFunc
}
