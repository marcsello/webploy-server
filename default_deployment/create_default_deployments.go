package default_deployment

import (
	"bytes"
	_ "embed"
	"fmt"
	"go.uber.org/zap"
	"webploy-server/site"
)

const SystemCreatorName = "_system"

//go:embed index.html
var defaultDeploymentIndexContent

func CreateDefaultDeploymentsForSites(sitesProvider site.Provider, lgr *zap.Logger) error {
	lgr.Debug("Deploying default content for newly created sites...", zap.Strings("newSites", sitesProvider.GetNewSiteNamesSinceInit()))

	for _, siteName := range sitesProvider.GetNewSiteNamesSinceInit() {

		s, ok := sitesProvider.GetSite(siteName)
		if !ok {
			return fmt.Errorf("site does not exist")
		}

		d, err := s.CreateNewDeployment(SystemCreatorName)
		if err != nil {
			return err
		}

		lgr.Debug("Creating default deployment", zap.String("siteName", siteName), zap.String("deploymentID", d.ID()))

		// Add the default stuff
		err = d.AddFile("index.html", bytes.NewReader(defaultDeploymentIndexContent))
		if err != nil {
			return err
		}

		// Finish deployment
		err = d.Finish()
		if err != nil {
			return err
		}

		// Set as live
		err = s.SetLiveDeploymentID(d.ID())
		if err != nil {
			return err
		}

	}

	return nil

}
