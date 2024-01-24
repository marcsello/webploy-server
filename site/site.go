package site

import (
	"webploy-server/config"
	"webploy-server/deployment"
)

// Site is an interface for a site
type Site interface {
	GetName() string
	GetConfig() config.SiteConfig
	ListDeploymentIDs() ([]string, error)
	GetDeployment(id string) (deployment.Deployment, error)
	CreateNewDeployment(creator string) (string, deployment.Deployment, error)
	DeleteDeployment(id string) error
	SetLiveDeploymentID(id string) error
	GetLiveDeploymentID() (string, error)
}
