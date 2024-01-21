package deployment

import "errors"

var ErrDeploymentAlreadyExists = errors.New("deployment already exists")
var ErrDeploymentNotExist = errors.New("deployment does not exist")
var ErrDeploymentFinished = errors.New("deployment finished")
var ErrDeploymentNotFinished = errors.New("deployment not finished")
