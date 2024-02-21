package site

import "errors"

var ErrSiteNameInvalid = errors.New("site name invalid")
var ErrDeploymentNotExists = errors.New("deployment not exists")
var ErrDeploymentExists = errors.New("deployment exists")
var ErrInvalidID = errors.New("invalid id")
var ErrDeploymentLive = errors.New("deployment is live")
var ErrDeploymentNotFinished = errors.New("deployment not finished")
