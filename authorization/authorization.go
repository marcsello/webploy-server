package authorization

import (
	_ "embed"
	"webploy-server/config"
)

func InitAuthorizator(cfg config.AuthorizationProviderConfig) (Provider, error) {
	return NewCasbinProvider(cfg.PolicyFile)
}
