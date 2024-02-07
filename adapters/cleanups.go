package adapters

import (
	"go.uber.org/zap"
	"time"
	"webploy-server/deployment"
	"webploy-server/site"
)

func DeleteOldDeployments(s site.Site, logger *zap.Logger) (int, error) {

	// TODO: Disable cleanup?

	var oldDeployments []string

	err := s.IterDeployments(func(id string, d deployment.Deployment, isLive bool) (bool, error) {
		var err error
		l := logger.With(zap.String("deploymentID", id))

		if isLive {
			l.Debug("Ignoring live deployment")
			return true, nil // continue, ignore live deployment
		}

		var finished bool
		finished, err = d.IsFinished()
		if err != nil {
			l.Error("Could not read finished status", zap.Error(err))
			return false, err // break
		}

		if !finished {
			logger.Debug("Deployment is not finished, ignoring from old cleanup")
			return true, nil // continue
		}

		oldDeployments = append(oldDeployments, id)

		return true, nil // continue
	})
	if err != nil {
		logger.Error("Something went wrong wile gathering deployments for deletion", zap.Error(err))
		return 0, err
	}

	// check if the limit has been hit
	if uint(len(oldDeployments)) <= s.GetConfig().MaxHistory {
		logger.Debug("MaxHistory limit hasn't reached yet for this deployment. Nothing to do.")
		return 0, nil
	}

	var deploymentsToDelete []string

	// TODO: finish this: find and put the sites intended for deletion in the list

	logger.Debug("Gathered old deployments for deletion", zap.Strings("deploymentsToDelete", deploymentsToDelete))

	for i, id := range deploymentsToDelete {
		logger.Info("Deleting old deployment", zap.String("deploymentID", id), zap.Int("i", i))
		err = s.DeleteDeployment(id)
		if err != nil {
			logger.Error("Error while deleting old deployment", zap.Error(err))
			return i, err
		}
	}

	return len(deploymentsToDelete), nil
}

func DeleteStaleDeployments(s site.Site, referenceNow time.Time, logger *zap.Logger) (int, error) {

	if s.GetConfig().StaleCleanupTimeout == 0 {
		logger.Debug("Stale cleanup disabled for this site. Nothing to do....")
		return 0, nil
	}

	var deploymentsToDelete []string

	err := s.IterDeployments(func(id string, d deployment.Deployment, isLive bool) (bool, error) {
		var err error
		l := logger.With(zap.String("deploymentID", id))

		if isLive {
			l.Debug("Ignoring live deployment")
			return true, nil // continue, ignore live deployment
		}

		var finished bool
		finished, err = d.IsFinished()
		if err != nil {
			l.Error("Could not read finished status", zap.Error(err))
			return false, err // break
		}

		if finished {
			logger.Debug("Deployment is finished, ignoring from stale cleanup")
			return true, nil // continue
		}

		var lastActivity time.Time
		lastActivity, err = d.LastActivity()
		if err != nil {
			l.Error("Could not read last activity", zap.Error(err))
			return false, err // break
		}

		inactiveSince := referenceNow.Sub(lastActivity)
		shouldBeCleaned := inactiveSince > s.GetConfig().StaleCleanupTimeout
		l.Debug("Figured out current inactivity time",
			zap.Duration("inactiveSince", inactiveSince),
			zap.Bool("shouldBeCleaned", shouldBeCleaned),
		)

		if shouldBeCleaned {
			deploymentsToDelete = append(deploymentsToDelete, id)
		}

		return true, nil // continue
	})
	if err != nil {
		logger.Error("Something went wrong wile gathering deployments for deletion", zap.Error(err))
		return 0, err
	}

	logger.Debug("Gathered stale deployments for deletion", zap.Strings("deploymentsToDelete", deploymentsToDelete))

	for i, id := range deploymentsToDelete {
		logger.Info("Deleting stale deployment", zap.String("deploymentID", id), zap.Int("i", i))
		err = s.DeleteDeployment(id)
		if err != nil {
			logger.Error("Error while deleting stale deployment", zap.Error(err))
			return i, err
		}
	}

	return len(deploymentsToDelete), nil
}
