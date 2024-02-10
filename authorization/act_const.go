package authorization

const (
	// ActCreateDeployment - Ability to create a new deployment
	ActCreateDeployment = "create-deployment"

	// ActUploadSelf ability to upload files into a deployment created by the current user
	ActUploadSelf = "upload-self"

	// ActUploadAny ability to upload files into a deployment created by any user
	ActUploadAny = "upload-any"

	// ActFinishSelf ability to finish a deployment that is created by the current user
	ActFinishSelf = "finish-self"

	// ActFinishAny ability to finish any deployment created by any user
	ActFinishAny = "finish-any"

	// ActReadLive ability to read information of the current live deployment for a site
	ActReadLive = "read-live"

	// ActUpdateLive ability to update the current live deployment for a site to any of the uploaded, finished and not deleted deployments
	ActUpdateLive = "update-live"

	// ActListDeployments ability to list available deployments of a site
	ActListDeployments = "list-deployments"

	// ActReadDeployment ability to read information of any deployment
	ActReadDeployment = "read-deployment"
)
