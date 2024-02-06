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
const validDeploymentIDKey = "deployment_id"
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

		completedRequestFields := []zapcore.Field{
			zap.Int("status", ctx.Writer.Status()),
			zap.Duration("latency", latency),
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
		l := GetLoggerFromContext(ctx)
		siteName := ctx.Param("siteName")
		s, ok := siteProvider.GetSite(siteName)
		if !ok {
			l.Warn("Trying to access a non-existing site")
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		ctx.Set(loggerKey, l.With(zap.String("site", s.GetName()))) // add site name to the logger
		ctx.Set(validSiteKey, s)
	}
}

func GetSiteFromContext(ctx *gin.Context) site.Site {
	val, ok := ctx.Get(validSiteKey)
	if !ok {
		panic("could not read site from context")
	}
	var s site.Site
	s, ok = val.(site.Site)
	if !ok {
		panic("could not cast site to site type")
	}
	return s
}

func validDeploymentMiddleware() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		s := GetSiteFromContext(ctx)
		l := GetLoggerFromContext(ctx)

		deploymentID := ctx.Param("deploymentID")
		d, err := s.GetDeployment(deploymentID)
		if err != nil {
			GetLoggerFromContext(ctx).Debug("Trying to access a non-existing deployment")
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		ctx.Set(loggerKey, l.With(zap.String("deploymentID", deploymentID))) // add deployment id to the logger
		ctx.Set(validDeploymentKey, d)
		ctx.Set(validDeploymentIDKey, deploymentID)
	}
}

func GetDeploymentFromContext(ctx *gin.Context) (string, deployment.Deployment) {
	val, ok := ctx.Get(validDeploymentKey)
	if !ok {
		panic("could not read deployment from context")
	}

	var d deployment.Deployment
	d, ok = val.(deployment.Deployment)
	if !ok {
		panic("could not cast deployment to deployment type")
	}

	val, ok = ctx.Get(validDeploymentIDKey)
	if !ok {
		panic("could not read deployment id context")
	}

	var id string
	id, ok = val.(string)
	if !ok {
		panic("could not cast deployment id to string")
	}

	return id, d
}

// injectUsernameToLogger should be only added after both authN and logger middleware were invoked, it only updates the stored logger to include the username
func injectUsernameToLogger(ctx *gin.Context) {
	l := GetLoggerFromContext(ctx)

	username, ok := authentication.GetAuthenticatedUser(ctx)
	if ok {
		ctx.Set(loggerKey, l.With(zap.String("username", username))) // add username to the logger
	}

}
