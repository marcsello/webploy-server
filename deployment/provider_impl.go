package deployment

import (
	"go.uber.org/zap"
	"sync"
	"webploy-server/config"
	"webploy-server/utils"
)

type ProviderImpl struct {
	siteRoot   string // This is only used for sanity checks... maybe
	siteConfig config.SiteConfig
	mutex      *sync.RWMutex // This is only used to prevent loading from a deployment folder that is not yet initialized... pretty weak, as it don't protect when an initialization is not finished due to crashing
	logger     *zap.Logger
}

func InitDeploymentProvider(siteRoot string, siteConfig config.SiteConfig, lgr *zap.Logger) (Provider, error) {
	return &ProviderImpl{
		siteRoot:   siteRoot,
		siteConfig: siteConfig,
		mutex:      &sync.RWMutex{},
		logger:     lgr,
	}, nil
}

func (p *ProviderImpl) LoadDeployment(deploymentDir string) (Deployment, error) {

	subdir, err := utils.IsSubDir(p.siteRoot, deploymentDir)
	if err != nil {
		return nil, err
	}
	if !subdir {
		return nil, ErrDeploymentInvalidPath
	}

	p.mutex.RLock()
	defer p.mutex.RUnlock()

	var exists bool
	exists, err = utils.ExistsAndDirectory(deploymentDir)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrDeploymentDirectoryMissing
	}

	return &DeploymentImpl{
		infoProvider: nil, // <- TODO
		fullPath:     deploymentDir,
		siteConfig:   p.siteConfig,
	}, nil
}

func (p *ProviderImpl) InitDeployment(deploymentDir, creator string) (Deployment, error) {

	subdir, err := utils.IsSubDir(p.siteRoot, deploymentDir)
	if err != nil {
		return nil, err
	}
	if !subdir {
		return nil, ErrDeploymentInvalidPath
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	var exists bool
	exists, err = utils.ExistsAndDirectory(deploymentDir)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrDeploymentDirectoryMissing
	}

	return &DeploymentImpl{
		infoProvider: nil, // <- TODO
		fullPath:     deploymentDir,
		siteConfig:   p.siteConfig,
	}, nil

}
