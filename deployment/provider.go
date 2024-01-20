package deployment

import "webploy-server/config"

type Provider interface {
	LoadExistingDeployment(fullPath, id string, siteConfig config.SiteConfig) (Deployment, error)
	CreateNewDeployment(fullPath, id string, siteConfig config.SiteConfig, creator string) (Deployment, error)
}
