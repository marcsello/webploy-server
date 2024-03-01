package main

import (
	"github.com/gin-gonic/gin"
	"github.com/marcsello/webploy-server/api"
	"github.com/marcsello/webploy-server/authentication"
	"github.com/marcsello/webploy-server/authorization"
	"github.com/marcsello/webploy-server/config"
	"github.com/marcsello/webploy-server/default_deployment"
	"github.com/marcsello/webploy-server/hooks"
	"github.com/marcsello/webploy-server/jobs"
	"github.com/marcsello/webploy-server/site"
	"github.com/marcsello/webploy-server/utils"
	"gitlab.com/MikeTTh/env"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

var (
	// injected build time
	version        string
	commitHash     string
	buildTimestamp string
)

func main() {
	debug := env.Bool("WEBPLOY_DEBUG", false)

	var lgr *zap.Logger
	if debug {
		lgr = zap.Must(zap.NewDevelopment())
		gin.SetMode(gin.DebugMode)
		lgr.Warn("RUNNING IN DEBUG MODE!")
	} else {
		// TODO: read more logger options from config file???
		lgr = zap.Must(zap.NewProduction())
		gin.SetMode(gin.ReleaseMode)
	}
	defer lgr.Sync()

	lgr.Info("Starting webploy server...", zap.String("version", version), zap.String("commitHash", commitHash), zap.String("buildTimestamp", buildTimestamp))

	cfg, err := config.LoadConfig(lgr)
	if err != nil {
		lgr.Panic("Failed to load config", zap.Error(err))
	}

	lgr.Info("Initializing sites provider...")
	var sitesProvider site.Provider
	sitesProvider, err = site.InitSites(cfg.Sites, lgr)
	if err != nil {
		lgr.Panic("Failed to initialize sites provider", zap.Error(err))
	}

	go func() { // do it in the "background"
		err = default_deployment.CreateDefaultDeploymentsForSites(sitesProvider, lgr)
		if err != nil {
			lgr.Panic("Failed to create default deployments", zap.Error(err))
		}
	}()

	lgr.Info("Initializing hooks...")
	hooks.InitHooks(lgr)

	lgr.Info("Initializing authentication provider...")
	var authNProvider authentication.Provider
	authNProvider, err = authentication.InitAuthenticator(cfg.Authentication)
	if err != nil {
		lgr.Panic("Failed to initialize authentication provider", zap.Error(err))
	}

	lgr.Info("Initializing authorization provider...")
	var authZProvider authorization.Provider
	authZProvider, err = authorization.InitAuthorizator(cfg.Authorization, lgr)
	if err != nil {
		lgr.Panic("Failed to initialize authorization provider", zap.Error(err))
	}

	lgr.Info("Initializing API...")
	var apiDaemon utils.Daemon
	apiDaemon, err = api.InitApi(cfg.Listen, authNProvider, authZProvider, sitesProvider, lgr)
	if err != nil {
		lgr.Panic("Failed to initialize API", zap.Error(err))
	}

	lgr.Info("Initializing Job runner...")
	var jobRunnerDaemon utils.Daemon
	jobRunnerDaemon, err = jobs.InitJobRunner(lgr, sitesProvider)
	if err != nil {
		lgr.Panic("Failed to initialize job runner", zap.Error(err))
	}

	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, syscall.SIGINT)
	signal.Notify(stopSignal, syscall.SIGTERM)

	// all initialization done, start stuff
	lgr.Debug("Starting API...")
	err = apiDaemon.Start()
	if err != nil {
		lgr.Panic("Failed to start API", zap.Error(err))
	}

	lgr.Debug("Starting job runner...")
	err = jobRunnerDaemon.Start()
	if err != nil {
		lgr.Panic("Failed to start job runner", zap.Error(err))
	}

	lgr.Info("Ready!")
	select {
	case sig := <-stopSignal: // wait for stop signal, and stop gracefully
		lgr.Info("Stop signal recieved, stopping server...", zap.String("signal", sig.String()))
	case err = <-jobRunnerDaemon.ErrChan():
		lgr.Panic("Job runner daemon ran into a problem", zap.Error(err))
	case err = <-apiDaemon.ErrChan():
		lgr.Panic("API daemon ran into a problem", zap.Error(err))
	}

	lgr.Info("Stopping job runner...")
	err = jobRunnerDaemon.Destroy()
	if err != nil {
		lgr.Panic("Failed to destroy job runner", zap.Error(err))
	}

	lgr.Info("Stopping API...")
	err = apiDaemon.Destroy()
	if err != nil {
		lgr.Panic("Failed to destroy API", zap.Error(err))
	}

	lgr.Debug("Bye!")
}
