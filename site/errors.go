package site

import "errors"

var ErrInvalidID = errors.New("invalid id")
var ErrDeploymentLive = errors.New("deployment is live")
