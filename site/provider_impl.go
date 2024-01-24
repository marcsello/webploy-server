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
	sites        map[string]*SiteImpl // will be read only, no need for locking
	newSiteNames []string
}

func InitSites(cfg config.SitesConfig, lgr *zap.Logger) (Provider, error) {

	var firstTimers []string // names of sites that are just created

	sites := make(map[string]*SiteImpl, len(cfg.Sites))
	for _, siteCfg := range cfg.Sites {
		lgr.Info("Loading site", zap.String("Name", siteCfg.Name))

		// check for duplicate
		_, duplicate := sites[siteCfg.Name]
		if duplicate {
			return nil, fmt.Errorf("duplicate site config: %s", siteCfg.Name)
		}

		// figure out the path for site's files
		// typically /var/www/some_site
		fullPath := path.Join(cfg.Root, siteCfg.Name)
		lgr.Debug("Full path for site", zap.String("siteName", siteCfg.Name), zap.String("fullPath", fullPath))

		siteLogger := lgr.With(zap.String("siteName", siteCfg.Name))

		// initialize deployment provider for the site
		dp, err := deployment.InitDeploymentProvider(fullPath, siteCfg, siteLogger)
		if err != nil {
			lgr.Error("Failed to initialize deployment provider for site", zap.String("siteName", siteCfg.Name), zap.Error(err))
			return nil, err
		}
		lgr.Debug("Deployment provider successfully initialized", zap.String("siteName", siteCfg.Name))

		// create site object
		site := &SiteImpl{
			fullPath:           fullPath,
			deploymentsMutex:   sync.RWMutex{},
			cfg:                siteCfg,
			deploymentProvider: dp,
			logger:             siteLogger,
		}

		// run init stuff (create dir if needed)
		var first bool
		first, err = site.Init()
		if err != nil {
			return nil, err
		}
		if first {
			lgr.Debug("site is created as new", zap.String("siteName", siteCfg.Name))
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
