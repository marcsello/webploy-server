package site

import (
	"fmt"
	"go.uber.org/zap"
	"path"
	"sync"
	"webploy-server/config"
	"webploy-server/deployment"
)

type ProviderImpl struct {
	sites        map[string]*SiteImpl
	newSiteNames []string
}

func InitSites(cfg config.SitesConfig, lgr *zap.Logger, deploymentProvider deployment.Provider) (Provider, error) {

	var firstTimers []string // names of sites that are just created

	sites := make(map[string]*SiteImpl, len(cfg.Sites))
	for _, siteCfg := range cfg.Sites {
		lgr.Info("Loading site", zap.String("Name", siteCfg.Name))

		// check for duplicate
		_, duplicate := sites[siteCfg.Name]
		if duplicate {
			return nil, fmt.Errorf("duplicate site config: %s", siteCfg.Name)
		}

		// create site object
		site := &SiteImpl{
			fullPath:           path.Join(cfg.Root, siteCfg.Name),
			deploymentsMutex:   sync.RWMutex{},
			cfg:                siteCfg,
			deploymentProvider: deploymentProvider,
		}

		// run init stuff (create dir if needed)
		first, err := site.Init()
		if err != nil {
			return nil, err
		}
		if first {
			firstTimers = append(firstTimers, siteCfg.Name)
		}

		// store it
		sites[siteCfg.Name] = site
	}

	if len(sites) == 0 {
		lgr.Warn("No sites configured")
	}

	return &ProviderImpl{
		sites:        sites,
		newSiteNames: firstTimers,
	}, nil
}

func (p *ProviderImpl) GetSite(name string) (Site, bool) {
	site, ok := p.sites[name]
	return site, ok
}

func (p *ProviderImpl) GetNewSiteNamesSinceInit() []string {
	return p.newSiteNames
}
