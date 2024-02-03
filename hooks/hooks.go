package hooks

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"os/exec"
	"webploy-server/deployment"
	"webploy-server/site"
)

var logger *zap.Logger

func compileEnv(hook HookID, user string, d deployment.Deployment, dID string, altMeta string, s site.Site) ([]string, error) {
	envvars := []string{
		fmt.Sprintf("WEBPLOY_HOOK=%s", hook),
		fmt.Sprintf("WEBPLOY_USER=%s", user),
	}

	if s != nil {
		currentLive, err := s.GetLiveDeploymentID()
		if err != nil {
			return nil, err
		}

		envvars = append(envvars,
			fmt.Sprintf("WEBPLOY_SITE=%s", s.GetName()),
			fmt.Sprintf("WEBPLOY_SITE_PATH=%s", s.GetPath()),
			fmt.Sprintf("WEBPLOY_SITE_CURRENT_LIVE=%s", currentLive),
		)
	}

	if d != nil {
		i, err := d.GetFullInfo()
		if err != nil {
			return nil, err
		}

		envvars = append(envvars, fmt.Sprintf("WEBPLOY_DEPLOYMENT_CREATOR=%s", i.Creator))
		envvars = append(envvars, fmt.Sprintf("WEBPLOY_DEPLOYMENT_META=%s", i.Meta))
		envvars = append(envvars, fmt.Sprintf("WEBPLOY_DEPLOYMENT_PATH=%s", d.GetPath()))
	} else { // when the deployment is not yet created
		envvars = append(envvars, fmt.Sprintf("WEBPLOY_DEPLOYMENT_META=%s", altMeta))
	}
	if dID != "" {
		envvars = append(envvars, fmt.Sprintf("WEBPLOY_DEPLOYMENT_ID=%s", dID))
	}
	return envvars, nil
}

func RunHook(ctx context.Context, hook HookID, user string, d deployment.Deployment, dID string, altMeta string, s site.Site) (bool, error) { // <- TODO: this is a very stupid signature
	l := logger.With(zap.String("hook", string(hook)), zap.String("deploymentID", dID), zap.String("site", s.GetName()))

	hookPath := getHookPathFromConfig(hook, s.GetConfig().Hooks)
	if hookPath == "" {
		l.Debug("no hook configured")
		return true, nil
	}
	l = l.With(zap.String("hookPath", hookPath))

	extraEnv, err := compileEnv(hook, user, d, dID, altMeta, s)
	if err != nil {
		l.Error("Failed to compile envvars", zap.Error(err))
		return false, err
	}

	args := []string{string(hook)}
	if d != nil {
		args = append(args, d.GetPath())
	}

	x := exec.CommandContext(ctx, hookPath, args...)
	x.Env = append(x.Environ(), extraEnv...)

	var exitCode int
	var output []byte
	execOk := true

	output, err = x.CombinedOutput()
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
