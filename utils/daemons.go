package utils

type Daemon interface {
	// Start should be non-blocking in all cases, and it should start the stuff in a separate goroutine
	Start() error
	// Destroy implies that the "daemon" will be considered destroyed and will not be attempted to restart
	Destroy() error

	// ErrChan is a channel on which the daemon can spit out errors
	ErrChan() <-chan error
}
