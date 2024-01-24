package api

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"webploy-server/authentication"
	"webploy-server/authorization"
	"webploy-server/config"
	"webploy-server/site"
)

func InitApi(cfg config.ListenConfig, authNProvider authentication.Provider, authZProvider authorization.Provider, siteProvider site.Provider, lgr *zap.Logger) func() error {

	r := gin.New()
	r.Use(authNProvider.NewMiddleware()) // this also saves the username in the context

	siteGroup := r.Group("sites/:siteName")
	siteGroup.Use(ValidSiteMiddleware(siteProvider)) // this also saves the siteGroup in the context

	currentDeploymentGroup := siteGroup.Group("current")
	currentDeploymentGroup.GET("", authZProvider.NewMiddleware("read-current"), readCurrentDeployment)
	currentDeploymentGroup.PUT("", authZProvider.NewMiddleware("update-current"), updateCurrentDeployment)

	siteDeploymentsGroup := siteGroup.Group("deployments")
	siteDeploymentsGroup.GET("", authZProvider.NewMiddleware("list-deployments"), listDeployments)
	siteDeploymentsGroup.POST("", authZProvider.NewMiddleware("create-deployment"), createDeployment)
	siteDeploymentsGroup.POST(":deploymentID/upload", authZProvider.NewMiddleware("create-deployment"), ValidDeploymentMiddleware(), uploadToDeployment)
	siteDeploymentsGroup.POST(":deploymentID/finish", authZProvider.NewMiddleware("create-deployment"), ValidDeploymentMiddleware(), finishDeployment)

	return func() error {
		lgr.Info("Starting API server", zap.String("bind", cfg.BindAddr), zap.Bool("EnableTLS", cfg.EnableTLS))

		if cfg.EnableTLS {
			return r.RunTLS(cfg.BindAddr, cfg.TLSCert, cfg.TLSKey)
		} else {
			lgr.Warn("Running in HTTP mode without TLS")
			return r.Run(cfg.BindAddr)
		}
	}

}
