package authorization

import (
	_ "embed"
	"github.com/gin-gonic/gin"
	"webploy-server/config"
)

// Provider is an Authorization provider
type Provider interface {
	NewMiddleware(act string) gin.HandlerFunc
}

//go:embed model.conf
var modelConfigString

func InitAuthorizator(cfg config.AuthorizationProviderConfig) (Provider, error) {
	return NewCasbinProvider(cfg.PolicyFile)
}
