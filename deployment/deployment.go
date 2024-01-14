package deployment

import "webploy-server/config"

type Deployment interface {
	ID() string
	IsFinished() bool
	Delete()
}

type Provider interface {
	LoadExistingDeployment(fullPath, id string, siteConfig config.SiteConfig) (Deployment, error)
	CreateNewDeployment(fullPath, id string, siteConfig config.SiteConfig, creator string) (Deployment, error)
}
