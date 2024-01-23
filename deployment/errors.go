package deployment

import "errors"

var ErrDeploymentDirectoryMissing = errors.New("deployment directory is missing")
var ErrDeploymentInvalidPath = errors.New("deployment path is invalid")

var ErrDeploymentFinished = errors.New("deployment finished")
var ErrDeploymentNotFinished = errors.New("deployment not finished")
