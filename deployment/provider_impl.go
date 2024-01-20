package deployment

import (
	"os"
	"webploy-server/config"
	"webploy-server/utils"
)

type ProviderImpl struct {
}

func InitDeployments() (Provider, error) {
	return &ProviderImpl{}, nil
}

func (p *ProviderImpl) LoadExistingDeployment(fullPath, id string, siteConfig config.SiteConfig) (Deployment, error) {

	exists, err := utils.ExistsAndDirectory(fullPath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrDeploymentNotExist
	}

	// TODO

}

func (p *ProviderImpl) CreateNewDeployment(fullPath, id string, siteConfig config.SiteConfig, creator string) (Deployment, error) {

	err := os.Mkdir(fullPath, 0o750)
	if err != nil {
		if os.IsExist(err) {
			return nil, ErrDeploymentAlreadyExists
		}
		return nil, err
	}

	// TODO

}
