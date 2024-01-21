package deployment

import (
	"go.uber.org/zap"
	"os"
	"path"
	"sync"
	"webploy-server/config"
	"webploy-server/utils"
)

type ProviderImpl struct {
	siteRoot   string // Each Site have their own deployment provider instance, DeploymentProviders are sort-of singletons
	siteConfig config.SiteConfig
	mutex      *sync.RWMutex // Sadly, creating a folder and saving/reading stuff in it can not be done as a single atomic operation, so we still need some locking
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

func (p *ProviderImpl) getPathForId(id string) string {
	return path.Join(p.siteRoot, id)
}

func (p *ProviderImpl) LoadDeployment(id string) (Deployment, error) {
	fullPath := p.getPathForId(id) // this is an idempotent option, we could do it before locking

	p.mutex.RLock()
	defer p.mutex.RUnlock()

	exists, err := utils.ExistsAndDirectory(fullPath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrDeploymentNotExist
	}

	// TODO

}

func (p *ProviderImpl) CreateDeployment(id string, creator string) (Deployment, error) {
	fullPath := p.getPathForId(id)

	p.mutex.Lock()
	defer p.mutex.Unlock()

	err := os.Mkdir(fullPath, 0o750)
	if err != nil {
		if os.IsExist(err) {
			return nil, ErrDeploymentAlreadyExists
		}
		return nil, err
	}

	// TODO

}

func (p *ProviderImpl) DeleteDeployment(id string) error {
	fullPath := p.getPathForId(id)
	underDeletePath := fullPath + ".delete"

	p.mutex.Lock()
	defer p.mutex.Unlock()

	// TODO: ensure no other operations are pending on the deployment (maybe use a wait group) or terminate pending uploads

	// first, we "atomically" move it out of the way, so subsequent get requests will fail with deployment not existing
	// this also breaks the symlink
	err := os.Rename(fullPath, underDeletePath)
	if err != nil {
		return err
	}

	// we then do removing in the background
	go func() {
		p.logger.Debug("deployment delete started in the background", zap.String("underDeletePath", underDeletePath))
		e := os.RemoveAll(underDeletePath)
		if e != nil {
			p.logger.Error("failed to delete deployment folder", zap.Error(e), zap.String("underDeletePath", underDeletePath))
		}
	}()

	return nil
}
