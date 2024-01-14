package default_deployment

import (
	"fmt"
	"go.uber.org/zap"
	"webploy-server/site"
)

func CreateDefaultDeploymentsForSites(siteNames []string, lgr *zap.Logger) error {

	for _, siteName := range siteNames {

		s, ok := site.GetSite(siteName)
		if !ok {
			return fmt.Errorf("site does not exists")
		}

		d, err := s.CreateNewDeployment()
		if err != nil {
			return err
		}

		lgr.Debug("Creating default deployment", zap.String("siteName", siteName), zap.String("deploymentID", d.ID()))

		// TODO: actually create a deployment

	}

	return nil

}
