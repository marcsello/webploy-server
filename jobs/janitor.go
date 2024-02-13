package jobs

import (
	"github.com/marcsello/webploy-server/adapters"
	"github.com/marcsello/webploy-server/site"
	"go.uber.org/zap"
	"time"
)

// the janitorJob collects all unfinished deployments that haven't been touched for the defined stale_cleanup_timeout time, and cleans them up
type janitorJob struct {
	sites site.Provider
}

func (jj *janitorJob) Run(logger *zap.Logger) {
	now := time.Now()
	siteNames := jj.sites.GetAllSiteNames()

	for _, siteName := range siteNames {
		sLogger := logger.With(zap.String("siteName", siteName))
		s, ok := jj.sites.GetSite(siteName)
		if !ok {
			sLogger.Error("Trying to access a non-existing site. Ignoring...")
			continue
		}

		cnt, err := adapters.DeleteStaleDeployments(s, now, sLogger)
		if err != nil {
			sLogger.Error("Failure running Stale Deployments deletion. Skipping site...", zap.Error(err))
			continue
		}
		sLogger.Debug("Stale deployment cleanup completed", zap.Int("cnt", cnt))

		cnt, err = adapters.DeleteOldDeployments(s, sLogger)
		if err != nil {
			sLogger.Error("Failure running Old Deployments deletion. Skipping site...", zap.Error(err))
			continue
		}
		sLogger.Debug("Old deployment cleanup completed", zap.Int("cnt", cnt))

	}

}
