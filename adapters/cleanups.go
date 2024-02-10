package adapters

import (
	"go.uber.org/zap"
	"sort"
	"time"
	"webploy-server/deployment"
	"webploy-server/deployment/info"
	"webploy-server/site"
)

type deploymentInfo struct {
	id string
	ts time.Time
}

type deploymentInfos []deploymentInfo

func (di deploymentInfos) AsIDs() []string {
	r := make([]string, len(di))
	for i, d := range di {
		r[i] = d.id
	}
	return r
}

func (di deploymentInfos) Len() int {
	return len(di)
}

func (di deploymentInfos) Less(i, j int) bool {
	return di[i].ts.Before(di[j].ts)
}

func (di deploymentInfos) Swap(i, j int) {
	di[i], di[j] = di[j], di[i]
}

func getDeletableDeployments(maxHistory uint, dis deploymentInfos) deploymentInfos {
	if maxHistory >= uint(dis.Len()) {
		// nothing to do
		return deploymentInfos{} // returning nil here would be the same probably
	}
	disSorted := make(deploymentInfos, len(dis))
	copy(disSorted, dis)
	sort.Sort(disSorted)
	throwAway := uint(len(dis)) - maxHistory // this should not be a problem, because the above check ensures that this results in a positive integer
	return disSorted[0:throwAway]
}

func DeleteOldDeployments(s site.Site, logger *zap.Logger) (int, error) {

	// TODO: Ability to disable cleanup?

	var oldDeployments deploymentInfos

	err := s.IterDeployments(func(id string, d deployment.Deployment, isLive bool) (bool, error) {
		var err error
		l := logger.With(zap.String("deploymentID", id))

		if isLive {
			l.Debug("Ignoring live deployment")
			return true, nil // continue, ignore live deployment
		}

		var i info.DeploymentInfo
		i, err = d.GetFullInfo()
		if err != nil {
			l.Error("Could not read info from deployment", zap.Error(err))
			return false, err // break
		}

		if !i.IsFinished() {
			logger.Debug("Deployment is not finished, ignoring from old cleanup")
			return true, nil // continue
		}

		oldDeployments = append(oldDeployments, deploymentInfo{
			id: id,
			ts: i.CreatedAt,
		})

		return true, nil // continue
	})
	if err != nil {
		logger.Error("Something went wrong wile gathering deployments for deletion", zap.Error(err))
		return 0, err
	}

	// get the list of deployments to be deleted
	deploymentsToDelete := getDeletableDeployments(s.GetConfig().MaxHistory, oldDeployments)
	if len(deploymentsToDelete) == 0 {
		logger.Debug("No deployments can be marked for deletion. Nothing to do.")
		return 0, nil
	}

	logger.Debug("Gathered old deployments for deletion", zap.Strings("deploymentsToDelete", deploymentsToDelete.AsIDs()))

	for i, di := range deploymentsToDelete {
		logger.Info("Deleting old deployment", zap.String("deploymentID", di.id), zap.Int("i", i), zap.Time("createdAt", di.ts))
		err = s.DeleteDeployment(di.id)
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
