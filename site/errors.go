package site

import "errors"

var ErrDeploymentNotExists = errors.New("deployment not exists")
var ErrDeploymentExists = errors.New("deployment exists")
var ErrInvalidID = errors.New("invalid id")
var ErrDeploymentLive = errors.New("deployment is live")
