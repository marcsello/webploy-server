package deployment

import "errors"

var ErrDeploymentDirectoryMissing = errors.New("deployment directory is missing")
var ErrDeploymentInvalidPath = errors.New("deployment path is invalid")
var ErrUploadInvalidPath = errors.New("upload path is invalid")

var ErrDeploymentFinished = errors.New("deployment finished")

var ErrTooManyConcurrentUploads = errors.New("too many concurrent uploads")
var ErrUploadPending = errors.New("upload is pending")
