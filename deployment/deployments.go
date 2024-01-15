package deployment

import "webploy-server/config"

type Provider interface {
	LoadExistingDeployment(fullPath, id string, siteConfig config.SiteConfig) (Deployment, error)
	CreateNewDeployment(fullPath, id string, siteConfig config.SiteConfig, creator string) (Deployment, error)
}

type ProviderImpl struct {
}

func InitDeployments() (Provider, error) {
	return &ProviderImpl{}, nil
}

func (p *ProviderImpl) LoadExistingDeployment(fullPath, id string, siteConfig config.SiteConfig) (Deployment, error) {

}

func (p *ProviderImpl) CreateNewDeployment(fullPath, id string, siteConfig config.SiteConfig, creator string) (Deployment, error) {

}
