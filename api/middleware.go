package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"time"
	"webploy-server/authentication"
	"webploy-server/deployment"
	"webploy-server/site"
)

const validSiteKey = "site"
const validDeploymentKey = "deployment"
const loggerKey = "lgr"

func goodLoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		// some evil middlewares may modify this value, so we store it
		path := ctx.Request.URL.Path

		subLogger := logger.With(
			zap.String("method", ctx.Request.Method),
			zap.String("path", path),
			zap.String("query", ctx.Request.URL.RawQuery),
			zap.String("ip", ctx.ClientIP()),
			zap.String("user-agent", ctx.Request.UserAgent()),
		)

		ctx.Set(loggerKey, subLogger)

		ctx.Next() // <- execute next thing in the chain
		end := time.Now()

		latency := end.Sub(start)

		authUser, authOk := authentication.GetAuthenticatedUser(ctx)
		completedRequestFields := []zapcore.Field{
			zap.Int("status", ctx.Writer.Status()),
			zap.Duration("latency", latency),
			zap.Bool("authOk", authOk),
		}
		if authOk {
			completedRequestFields = append(completedRequestFields, zap.String("authUser", authUser))
		}

		if len(ctx.Errors) > 0 {
			// Append error field if this is an erroneous request.
			for _, e := range ctx.Errors.Errors() {
				subLogger.Error(e, completedRequestFields...)
			}
		}

		subLogger.Info(fmt.Sprintf("%s %s served: %d", ctx.Request.Method, path, ctx.Writer.Status()), completedRequestFields...) // <- always print this

	}
}

func GetLoggerFromContext(ctx *gin.Context) *zap.Logger { // This one panics
	var logger *zap.Logger
	l, ok := ctx.Get(loggerKey)
	if !ok {
		panic("could not access logger")
	}
	logger = l.(*zap.Logger)
	return logger
}

func validSiteMiddleware(siteProvider site.Provider) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		siteName := ctx.Param("siteName")
		s, ok := siteProvider.GetSite(siteName)
		if !ok {
			GetLoggerFromContext(ctx).Debug("Trying to access a non-existing site")
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		ctx.Set(validSiteKey, s)
	}
}

func GetSiteFromContext(ctx *gin.Context) (site.Site, bool) {
	val, ok := ctx.Get(validSiteKey)
	if !ok {
		return nil, false
	}
	var s site.Site
	s, ok = val.(site.Site)
	return s, ok
}

func validDeploymentMiddleware() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		s, ok := GetSiteFromContext(ctx)
		if !ok {
			GetLoggerFromContext(ctx).Debug("Trying to access a non-existing site")
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		deploymentID := ctx.Param("deploymentID")
		d, err := s.GetDeployment(deploymentID)
		if err != nil {
			GetLoggerFromContext(ctx).Debug("Trying to access a non-existing deployment")
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		ctx.Set(validDeploymentKey, d)
	}
}

func GetDeploymentFromContext(ctx *gin.Context) (deployment.Deployment, bool) {
	val, ok := ctx.Get(validSiteKey)
	if !ok {
		return nil, false
	}
	var d deployment.Deployment
	d, ok = val.(deployment.Deployment)
	return d, ok
}
