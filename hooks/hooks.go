package hooks

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"os/exec"
	"webploy-server/config"
)

var logger *zap.Logger

func RunHook(ctx context.Context, hooksConfig config.HooksConfig, hook HookID, vars HookVars) (bool, error) {
	l := logger.With(zap.String("hook", string(hook)), zap.String("deploymentID", vars.DeploymentID), zap.String("site", vars.SiteName))

	hookPath := getHookPathFromConfig(hook, hooksConfig)
	if hookPath == "" {
		l.Debug("no hook configured")
		return true, nil
	}
	l = l.With(zap.String("hookPath", hookPath))

	extraEnv := vars.compileEnvvars(hook)

	args := []string{string(hook)}
	if vars.DeploymentPath != "" {
		args = append(args, vars.DeploymentPath)
	}

	x := exec.CommandContext(ctx, hookPath, args...)
	x.Env = append(x.Environ(), extraEnv...)

	var exitCode int
	execOk := true

	output, err := x.CombinedOutput()
	if err != nil {
		var exiterr *exec.ExitError
		ok := errors.As(err, &exiterr)
		if ok {
			exitCode = exiterr.ExitCode()
			execOk = false
		} else {
			// some other error
			return false, err
		}
	}
	l.Info("Hook executed", zap.Int("exitCode", exitCode), zap.Bool("execOk", execOk), zap.ByteString("output", output))
	return execOk, nil
}

func InitHooks(lgr *zap.Logger) { // called from main on init
	logger = lgr
}
