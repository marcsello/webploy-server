package main

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/MikeTTh/env"
	"go.uber.org/zap"
	"webploy-server/api"
	"webploy-server/authentication"
	"webploy-server/authorization"
	"webploy-server/config"
	"webploy-server/default_deployment"
	"webploy-server/deployment"
	"webploy-server/site"
)

func main() {
	debug := env.Bool("WEBPLOY_DEBUG", false)

	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	var lgr *zap.Logger
	if debug {
		lgr = zap.Must(zap.NewDevelopment())
		gin.SetMode(gin.DebugMode)
		lgr.Warn("RUNNING IN DEBUG MODE!")
	} else {
		// TODO: read more logger options from config file
		lgr = zap.Must(zap.NewProduction())
		gin.SetMode(gin.ReleaseMode)
	}

	lgr.Info("Intializing deployment provider...")
	var deploymentProvider deployment.Provider
	deploymentProvider, err = deployment.InitDeployments()
	if err != nil {
		panic(err)
	}

	lgr.Info("Intializing sites provider...")
	var sitesProvider site.Provider
	sitesProvider, err = site.InitSites(cfg.Sites, lgr, deploymentProvider)
	if err != nil {
		panic(err)
	}

	err = default_deployment.CreateDefaultDeploymentsForSites(sitesProvider, lgr)
	if err != nil {
		panic(err)
	}

	lgr.Info("Initiating authentication provider...")
	var authNProvider authentication.Provider
	authNProvider, err = authentication.InitAuthenticator(cfg.Authentication)
	if err != nil {
		panic(err)
	}

	lgr.Info("Initiating authorization provider...")
	var authZProvider authorization.Provider
	authZProvider, err = authorization.InitAuthorizator(cfg.Authorization)
	if err != nil {
		panic(err)
	}

	lgr.Info("Initiating API...")
	run := api.InitApi(cfg.Listen, authNProvider, authZProvider, lgr)

	// run the api
	err = run()
	if err != nil {
		panic(err)
	}

}
