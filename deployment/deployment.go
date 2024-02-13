package deployment

import (
	"context"
	"github.com/marcsello/webploy-server/deployment/info"
	"io"
	"time"
)

type Deployment interface {
	GetPath() string
	AddFile(ctx context.Context, relpath string, stream io.ReadCloser) error
	IsFinished() (bool, error)
	Finish() error
	Creator() (string, error)
	LastActivity() (time.Time, error)
	GetFullInfo() (info.DeploymentInfo, error)
}
