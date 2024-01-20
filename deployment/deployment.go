package deployment

import (
	"io"
	"time"
)

type Deployment interface {
	ID() string
	AddFile(relpath string, stream io.Reader) error
	IsFinished() (bool, error)
	Finish() error
	Creator() (string, error)
	LastActivity() (time.Time, error)
	Delete() error
}
