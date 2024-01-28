package deployment

type Provider interface {
	InitDeployment(deploymentDir, creator, meta string) (Deployment, error) // Initializes a new deployment in an empty folder
	LoadDeployment(deploymentDir string) (Deployment, error)                // Load deployment from an already populated folder
}
