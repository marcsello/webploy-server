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

	lgr.Info("Intializing sites...")
	var newSites []string
	newSites, err = site.InitSites(cfg.Sites, lgr)

	lgr.Debug("Deploy default content for newly created sites...", zap.Strings("newSites", newSites))
	err = default_deployment.CreateDefaultDeploymentsForSites(newSites, lgr)
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
	var authRProvider authorization.Provider
	authRProvider, err = authorization.InitAuthorizator(cfg.Authorization)
	if err != nil {
		panic(err)
	}

	lgr.Info("Initiating API...")
	run := api.InitApi(cfg.Listen, authNProvider, authRProvider, lgr)

	// run the api
	err = run()
	if err != nil {
		panic(err)
	}

}
