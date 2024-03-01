package api

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/dyson/certman"
	limits "github.com/gin-contrib/size"
	"github.com/gin-gonic/gin"
	"github.com/marcsello/webploy-server/adapters"
	"github.com/marcsello/webploy-server/authentication"
	"github.com/marcsello/webploy-server/authorization"
	"github.com/marcsello/webploy-server/config"
	"github.com/marcsello/webploy-server/site"
	"github.com/marcsello/webploy-server/utils"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const DefaultRequestBodySize = 1024

type apiDaemon struct {
	errChan chan error
	lgr     *zap.Logger
	srv     *http.Server
	cm      *certman.CertMan
	tls     bool
}

func InitApi(cfg config.ListenConfig, authNProvider authentication.Provider, authZProvider authorization.Provider, siteProvider site.Provider, lgr *zap.Logger) (utils.Daemon, error) {

	r := gin.New()
	r.Use(goodLoggerMiddleware(lgr))     // <- This must be the first, other middlewares may use it... and funnily enough this maybe uses other middlewares as well
	r.Use(authNProvider.NewMiddleware()) // this also saves the username in the context (the username may be logged)
	r.Use(injectUsernameToLogger)        // This should be included after AuthN and logger middlewares, it simply loads the username from the context and adds it to the logger.

	siteGroup := r.Group("sites/:siteName")
	siteGroup.Use(validSiteMiddleware(siteProvider)) // this also saves the siteGroup in the context

	currentDeploymentGroup := siteGroup.Group("live")
	currentDeploymentGroup.GET("", authZProvider.NewMiddleware(authorization.ActReadLive), readLiveDeployment)
	currentDeploymentGroup.PUT("", limits.RequestSizeLimiter(DefaultRequestBodySize), authZProvider.NewMiddleware(authorization.ActUpdateLive), updateLiveDeployment)

	siteDeploymentsGroup := siteGroup.Group("deployments")
	siteDeploymentsGroup.GET("", authZProvider.NewMiddleware(authorization.ActListDeployments), listDeployments)
	siteDeploymentsGroup.POST("", limits.RequestSizeLimiter(DefaultRequestBodySize), authZProvider.NewMiddleware(authorization.ActCreateDeployment), createDeployment)

	siteDeploymentsGroup.GET(":deploymentID", authZProvider.NewMiddleware(authorization.ActReadDeployment), validDeploymentMiddleware(), readDeployment)
	siteDeploymentsGroup.DELETE(":deploymentID", authZProvider.NewMiddleware(), validDeploymentMiddleware(), deleteDeployment)

	siteDeploymentsGroup.POST(":deploymentID/upload", authZProvider.NewMiddleware(), validDeploymentMiddleware(), uploadFileToDeployment)
	siteDeploymentsGroup.POST(":deploymentID/uploadTar", authZProvider.NewMiddleware(), validDeploymentMiddleware(), uploadTarToDeployment)
	siteDeploymentsGroup.POST(":deploymentID/finish", limits.RequestSizeLimiter(DefaultRequestBodySize), authZProvider.NewMiddleware(), validDeploymentMiddleware(), finishDeployment)

	srv := &http.Server{
		Addr:              cfg.BindAddr,
		Handler:           r,
		ReadHeaderTimeout: 2 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	// setup cert hot reload
	var cm *certman.CertMan
	if cfg.EnableTLS {
		var err error
		cm, err = certman.New(cfg.TLSCert, cfg.TLSKey)
		if err != nil {
			return nil, err
		}
		cm.Logger(adapters.LogAdapter{L: lgr.With(zap.String("src", "certman"))})
		srv.TLSConfig = &tls.Config{
			GetCertificate: cm.GetCertificate,
			MinVersion:     tls.VersionTLS12,
		}
	}

	return &apiDaemon{
		errChan: make(chan error, 1),
		lgr:     lgr,
		srv:     srv,
		cm:      cm,
		tls:     cfg.EnableTLS,
	}, nil

}

func (ad *apiDaemon) Start() error {
	ad.lgr.Info("Starting API server", zap.String("bind", ad.srv.Addr), zap.Bool("EnableTLS", ad.tls))

	if ad.tls {
		e := ad.cm.Watch()
		if e != nil {
			return e
		}
		go func() {
			err := ad.srv.ListenAndServeTLS("", "")
			if !errors.Is(err, http.ErrServerClosed) {
				ad.errChan <- err
			}
		}()
		return nil

	} else {
		ad.lgr.Warn("Running in HTTP mode without TLS")
		go func() {
			err := ad.srv.ListenAndServe()
			if !errors.Is(err, http.ErrServerClosed) {
				ad.errChan <- err
			}
		}()
		return nil

	}
}

func (ad *apiDaemon) Destroy() error {
	if ad.tls {
		ad.cm.Stop()
	}
	return ad.srv.Shutdown(context.Background())
}

func (ad *apiDaemon) ErrChan() <-chan error {
	return ad.errChan
}
