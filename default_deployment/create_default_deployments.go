package default_deployment

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"go.uber.org/zap"
	"io"
	"webploy-server/site"
)

const SystemCreatorName = "_system"

//go:embed index.html
var defaultDeploymentIndexContent string

func CreateDefaultDeploymentsForSites(sitesProvider site.Provider, lgr *zap.Logger) error {
	lgr.Debug("Deploying default content for newly created sites...", zap.Strings("newSites", sitesProvider.GetNewSiteNamesSinceInit()))

	for _, siteName := range sitesProvider.GetNewSiteNamesSinceInit() {

		s, ok := sitesProvider.GetSite(siteName)
		if !ok {
			return fmt.Errorf("site does not exist")
		}

		id, d, err := s.CreateNewDeployment(SystemCreatorName, "")
		if err != nil {
			return err
		}

		lgr.Debug("Creating default deployment", zap.String("siteName", siteName), zap.String("deploymentID", id))

		// Add the default stuff
		err = d.AddFile(context.Background(), "index.html", io.NopCloser(bytes.NewReader([]byte(defaultDeploymentIndexContent))))
		if err != nil {
			return err
		}

		// Finish deployment
		err = d.Finish()
		if err != nil {
			return err
		}

		// Set as live
		err = s.SetLiveDeploymentID(id)
		if err != nil {
			return err
		}

	}

	return nil

}
