package hooks

import (
	"context"
	"go.uber.org/zap"
	"webploy-server/config"
)

var (
	logger *zap.Logger
	exc    Executor
)

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

	exitCode, output, err := exc(ctx, hookPath, args, extraEnv)
	if err != nil {
		l.Error("Error while executing hook", zap.Error(err))
		return false, err
	}

	l.Info("Hook executed successfully", zap.Int("exitCode", exitCode), zap.ByteString("output", output))
	return exitCode == 0, nil
}

func InitHooks(lgr *zap.Logger) { // called from main on init
	logger = lgr
	exc = DefaultExecutor // set the default executor
}
