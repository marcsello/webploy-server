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
	"gitlab.com/MikeTTh/env"
	"go.uber.org/zap"
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
	runApi := api.InitApi(cfg.Listen, authNProvider, authZProvider, sitesProvider, lgr)

	lgr.Info("Initializing Job runner...")
	var runJobs func() error
	runJobs, err = jobs.InitJobRunner(lgr, sitesProvider)
	if err != nil {
		lgr.Panic("Failed to initialize job runner", zap.Error(err))
	}

	// run jobs
	go func() {
		e := runJobs()
		if e != nil {
			lgr.Panic("Error running Jobs", zap.Error(e))
		}
	}()

	// run the api
	err = runApi()
	if err != nil {
		lgr.Panic("Error running API", zap.Error(err))
	}

}
