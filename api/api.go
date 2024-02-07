package api

import (
	limits "github.com/gin-contrib/size"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"webploy-server/authentication"
	"webploy-server/authorization"
	"webploy-server/config"
	"webploy-server/site"
)

const DefaultRequestBodySize = 1024

func InitApi(cfg config.ListenConfig, authNProvider authentication.Provider, authZProvider authorization.Provider, siteProvider site.Provider, lgr *zap.Logger) func() error {

	r := gin.New()
	r.Use(goodLoggerMiddleware(lgr))     // <- This must be the first, other middlewares may use it... and funnily enough this maybe uses other middlewares as well
	r.Use(authNProvider.NewMiddleware()) // this also saves the username in the context (the username may be logged)
	r.Use(injectUsernameToLogger)        // This should be included after AuthN and logger middlewares, it simply loads the username from the context and adds it to the logger.

	siteGroup := r.Group("sites/:siteName")
	siteGroup.Use(validSiteMiddleware(siteProvider)) // this also saves the siteGroup in the context

	currentDeploymentGroup := siteGroup.Group("live")
	currentDeploymentGroup.GET("", authZProvider.NewMiddleware("read-live"), readLiveDeployment)
	currentDeploymentGroup.PUT("", limits.RequestSizeLimiter(DefaultRequestBodySize), authZProvider.NewMiddleware("update-live"), updateLiveDeployment)

	siteDeploymentsGroup := siteGroup.Group("deployments")
	siteDeploymentsGroup.GET("", authZProvider.NewMiddleware("list-deployments"), listDeployments)
	siteDeploymentsGroup.GET(":deploymentID", authZProvider.NewMiddleware("read-deployment"), readDeployment)

	siteDeploymentsGroup.POST("", limits.RequestSizeLimiter(DefaultRequestBodySize), authZProvider.NewMiddleware("create-deployment"), createDeployment)
	siteDeploymentsGroup.POST(":deploymentID/upload", authZProvider.NewMiddleware(), validDeploymentMiddleware(), uploadToDeployment)
	siteDeploymentsGroup.POST(":deploymentID/finish", limits.RequestSizeLimiter(DefaultRequestBodySize), authZProvider.NewMiddleware(), validDeploymentMiddleware(), finishDeployment)

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
