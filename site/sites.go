package site

import (
	"fmt"
	"go.uber.org/zap"
	"path"
	"sync"
	"webploy-server/config"
)

var sites map[string]*Site

func InitSites(cfg config.SitesConfig, lgr *zap.Logger) ([]string, error) {

	var firstTimers []string // names of sites that are just created

	sites = make(map[string]*Site, len(cfg.Sites))
	for _, siteCfg := range cfg.Sites {
		lgr.Info("Loading site", zap.String("Name", siteCfg.Name))

		// check for duplicate
		_, duplicate := sites[siteCfg.Name]
		if duplicate {
			return nil, fmt.Errorf("duplicate site config: %s", siteCfg.Name)
		}

		// create site object
		site := &Site{
			fullPath:         path.Join(cfg.Root, siteCfg.Name),
			deploymentsMutex: sync.RWMutex{},
			cfg:              siteCfg,

			deploymentProvider: nil, // <- big TODO
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

	return firstTimers, nil
}

func GetSite(name string) (*Site, bool) {
	site, ok := sites[name]
	return site, ok
}
