package hooks

import (
	"context"
	"errors"
	"os/exec"
)

// Executor is basically a shim for os.exec for easier testing
type Executor func(ctx context.Context, name string, args []string, extraEnv []string) (int, []byte, error)

func DefaultExecutor(ctx context.Context, name string, args []string, extraEnv []string) (int, []byte, error) {
	x := exec.CommandContext(ctx, name, args...)
	x.Env = append(x.Environ(), extraEnv...)

	output, err := x.CombinedOutput()
	var exitCode int

	if err != nil {
		var exiterr *exec.ExitError
		ok := errors.As(err, &exiterr)
		if ok {
			exitCode = exiterr.ExitCode()
		} else {
			// some other error
			return 0, nil, err
		}
	}

	return exitCode, output, nil
}
