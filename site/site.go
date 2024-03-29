package site

import (
	"github.com/marcsello/webploy-server/config"
	"github.com/marcsello/webploy-server/deployment"
)

type DeploymentIterator func(id string, d deployment.Deployment, isLive bool) (cont bool, err error)

// Site is an interface for a site
type Site interface {
	GetName() string
	GetPath() string
	GetConfig() config.SiteConfig
	ListDeploymentIDs() ([]string, error)
	GetDeployment(id string) (deployment.Deployment, error)
	IterDeployments(iter DeploymentIterator) error
	CreateNewDeployment(creator, meta string) (string, deployment.Deployment, error)
	DeleteDeployment(id string) error
	SetLiveDeploymentID(id string) error
	GetLiveDeploymentID() (string, error)
}
