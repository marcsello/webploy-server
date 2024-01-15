package deployment

import (
	"io"
	"time"
	"webploy-server/config"
)

type Deployment interface {
	ID() string
	AddFile(relpath string, stream io.Reader) error
	IsFinished() bool
	Finish() error
	Creator() string
	LastActivity() time.Time
	Delete() error
}

type DeploymentImpl struct {
	state      StateProvider
	fullPath   string
	id         string
	siteConfig config.SiteConfig
}

func (d *DeploymentImpl) ID() string {
	return d.id
}

func (d *DeploymentImpl) IsFinished() bool {
	return d.state.IsFinished()
}

func (d *DeploymentImpl) Creator() string {
	return d.state.Creator()
}
