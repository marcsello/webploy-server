package api

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"webploy-server/authentication"
	"webploy-server/authorization"
	"webploy-server/config"
)

func InitApi(cfg config.ListenConfig, authNProvider authentication.Provider, authRProvider authorization.Provider, lgr *zap.Logger) func() error {

	r := gin.New()
	r.Use(authNProvider.NewMiddleware())

	site := r.Group("sites/:siteName")
	site.Use(ValidSiteMiddleware())

	currentDeployment := site.Group("current")
	currentDeployment.GET("current", authRProvider.NewMiddleware("read-current"), readCurrentDeployment)
	currentDeployment.PUT("current", authRProvider.NewMiddleware("update-current"), updateCurrentDeployment)

	siteDeployments := site.Group("deployments")
	siteDeployments.GET("", authRProvider.NewMiddleware("list-deployments"), listDeployments)
	siteDeployments.POST("", authRProvider.NewMiddleware("create-deployment"), createDeployment)
	siteDeployments.POST(":deploymentID/upload", authRProvider.NewMiddleware("create-deployment"), uploadToDeployment)
	siteDeployments.POST(":deploymentID/finish", authRProvider.NewMiddleware("create-deployment"), finishDeployment)

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
