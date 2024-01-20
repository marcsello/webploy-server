package deployment

import (
	"io"
	"time"
	"webploy-server/config"
	"webploy-server/deployment/info"
)

type DeploymentImpl struct {
	infoProvider info.InfoProvider // <- store all state info here, as the Deployment objects generally live for a single request and they are not shared
	fullPath     string
	id           string
	siteConfig   config.SiteConfig
}

func (d *DeploymentImpl) ID() string {
	return d.id
}

func (d *DeploymentImpl) IsFinished() (bool, error) {
	var isFinished bool
	err := d.infoProvider.Tx(true, func(i *info.DeploymentInfo) error {
		isFinished = i.IsFinished()
		return nil
	})
	return isFinished, err
}

func (d *DeploymentImpl) Creator() (string, error) {
	var creator string
	err := d.infoProvider.Tx(true, func(i *info.DeploymentInfo) error {
		creator = i.Creator
		return nil
	})
	return creator, err
}

func (d *DeploymentImpl) LastActivity() (time.Time, error) {
	var lastActivity time.Time
	err := d.infoProvider.Tx(true, func(i *info.DeploymentInfo) error {
		lastActivity = i.LastActivityAt
		return nil
	})
	return lastActivity, err
}

func (d *DeploymentImpl) Finish() error {
	// TODO: don't allow finishing while uploads are still pending
	// Store the pending uploads in the info?
	return d.infoProvider.Tx(false, func(i *info.DeploymentInfo) error {
		if i.IsDeleting() {
			return ErrDeploymentDeleting
		}

		if i.IsFinished() {
			return ErrDeploymentFinished
		}

		i.State = info.DeploymentStateFinished
		now := time.Now()
		i.FinishedAt = &now
		i.LastActivityAt = now

		return nil

	})
}

func (d *DeploymentImpl) AddFile(relpath string, stream io.Reader) error {
	err := d.infoProvider.Tx(false, func(i *info.DeploymentInfo) error {
		if i.IsDeleting() {
			return ErrDeploymentDeleting
		}

		if i.IsFinished() {
			return ErrDeploymentFinished
		}

		i.LastActivityAt = time.Now()

		return nil

	})

	// TODO: handle upload

}

func (d *DeploymentImpl) Delete() error {
	err := d.infoProvider.Tx(false, func(i *info.DeploymentInfo) error {
		if i.IsDeleting() {
			return ErrDeploymentDeleting
		}

		i.State = info.DeploymentStateDeleting
		i.LastActivityAt = time.Now()

		return nil
	})
	if err != nil {
		return err
	}

	// TODO: actual deletion

	return nil
}
