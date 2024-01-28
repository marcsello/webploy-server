package jobs

import (
	"go.uber.org/zap"
	"time"
	"webploy-server/site"
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
			sLogger.Error("Trying to access a non-existing site")
			continue
		}

		deploymentIDs, err := s.ListDeploymentIDs()
		if err != nil {
			sLogger.Error("Failed to list deployment IDs", zap.Error(err))
			continue
		}

		for _, deploymentID := range deploymentIDs {
			dLogger := sLogger.With(zap.String("deploymentID", deploymentID))
			deployment, e := s.GetDeployment(deploymentID)
			if e != nil {
				dLogger.Error("Could not access deployment", zap.Error(e))
				continue
			}

			var finished bool
			finished, e = deployment.IsFinished()
			if e != nil {
				dLogger.Error("Could not read finished status", zap.Error(e))
				continue
			}

			if finished {
				logger.Debug("Deployment is finished, ignoring from stale cleanup")
				continue // skip
			}

			var lastActivity time.Time
			lastActivity, e = deployment.LastActivity()
			if e != nil {
				dLogger.Error("Could not read finished status", zap.Error(e))
				continue
			}

			inactiveSince := now.Sub(lastActivity)
			shouldBeCleaned := inactiveSince > s.GetConfig().StaleCleanupTimeout
			dLogger.Debug("Figured out current inactivity time",
				zap.Duration("inactiveSince", inactiveSince),
				zap.Bool("shouldBeCleaned", shouldBeCleaned),
			)

			if shouldBeCleaned {
				dLogger.Info("Deleting stale deployment")
				e = s.DeleteDeployment(deploymentID)

				// TODO: Terminate pending uploads to the deployment (if any)

				if e != nil {
					dLogger.Error("Failed to delete deployment", zap.Error(e))
				} else {
					dLogger.Debug("Deleted stale deployment")
				}
			}

		}

	}

}
