package authorization

import (
	_ "embed"
	"github.com/marcsello/webploy-server/config"
	"go.uber.org/zap"
)

func InitAuthorizator(cfg config.AuthorizationProviderConfig, logger *zap.Logger) (Provider, error) {
	return NewCasbinProvider(cfg.PolicyFile, logger)
}
