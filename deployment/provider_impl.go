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

func (p *ProviderImpl) createDeployment(deploymentDir string) (*DeploymentImpl, error) {
	// validate directory

	// make sure it's a subdir in the site folder
	subdir, err := utils.IsSubDir(p.siteRoot, deploymentDir)
	if err != nil {
		return nil, err
	}
	if !subdir {
		return nil, ErrDeploymentInvalidPath
	}

	// make sure it exists, and it is indeed a directory
	var exists bool
	exists, err = utils.ExistsAndDirectory(deploymentDir)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrDeploymentDirectoryMissing
	}

	// if all good, create the deployment
	return NewDeployment(deploymentDir, p.siteConfig, p.logger.With(zap.String("deploymentDir", deploymentDir))), nil

}

func (p *ProviderImpl) LoadDeployment(deploymentDir string) (Deployment, error) {
	p.mutex.RLock() // since we are not going to initialize this deployment, a read lock is enough
	defer p.mutex.RUnlock()

	return p.createDeployment(deploymentDir)
}

func (p *ProviderImpl) InitDeployment(deploymentDir, creator string) (Deployment, error) {
	p.mutex.Lock() // we will create some fundamental stuff, we need the write lock
	defer p.mutex.Unlock()

	d, err := p.createDeployment(deploymentDir)

	// initialize deployment, create... stuff, idk
	err = d.Init(creator)
	if err != nil {
		return nil, err
	}

	return d, nil
}
