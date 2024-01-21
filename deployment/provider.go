package deployment

type Provider interface {
	CreateDeployment(id string, creator string) (Deployment, error)
	LoadDeployment(id string) (Deployment, error)
	DeleteDeployment(id string) error
}
