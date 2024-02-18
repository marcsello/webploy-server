package default_deployment

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"github.com/marcsello/webploy-server/authentication"
	"github.com/marcsello/webploy-server/site"
	"go.uber.org/zap"
	"io"
)

const SystemCreatorName = authentication.SystemPrefix + "system"

//go:embed index.html
var defaultDeploymentIndexContent string

func CreateDefaultDeploymentsForSites(sitesProvider site.Provider, lgr *zap.Logger) error {
	lgr.Debug("Deploying default content for newly created sites...", zap.Strings("newSites", sitesProvider.GetNewSiteNamesSinceInit()))

	for _, siteName := range sitesProvider.GetNewSiteNamesSinceInit() {
		sLogger := lgr.With(zap.String("siteName", siteName))

		s, ok := sitesProvider.GetSite(siteName)
		if !ok {
			err := fmt.Errorf("site does not exist")
			sLogger.Error("Trying to access a non-existing site", zap.Error(err))
			return err
		}

		id, d, err := s.CreateNewDeployment(SystemCreatorName, "")
		if err != nil {
			sLogger.Error("Failure while creating the initial default deployment", zap.Error(err))
			return err
		}
		dLogger := sLogger.With(zap.String("deploymentID", id))

		dLogger.Debug("Created initial default deployment")

		// Add the default stuff
		err = d.AddFile(context.Background(), "index.html", io.NopCloser(bytes.NewReader([]byte(defaultDeploymentIndexContent))))
		if err != nil {
			dLogger.Error("Failure while adding the index file to the deployment", zap.Error(err))
			return err
		}

		// Finish deployment
		err = d.Finish()
		if err != nil {
			dLogger.Error("Failure while finishing the deployment", zap.Error(err))
			return err
		}

		// Set as live
		err = s.SetLiveDeploymentID(id)
		if err != nil {
			dLogger.Error("Failure while setting the deployment as live", zap.Error(err))
			return err
		}

		dLogger.Info("Created initial deployment for new site")
	}

	return nil

}
