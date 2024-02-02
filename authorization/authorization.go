package authorization

import (
	_ "embed"
	"go.uber.org/zap"
	"webploy-server/config"
)

func InitAuthorizator(cfg config.AuthorizationProviderConfig, logger *zap.Logger) (Provider, error) {
	return NewCasbinProvider(cfg.PolicyFile, logger)
}
