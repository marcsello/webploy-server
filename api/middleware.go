package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"webploy-server/deployment"
	"webploy-server/site"
)

const ValidSiteKey = "site"
const ValidDeploymentKey = "deployment"

func ValidSiteMiddleware(siteProvider site.Provider) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		siteName := ctx.Param("siteName")
		s, ok := siteProvider.GetSite(siteName)
		if !ok {
			// TODO: log
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		ctx.Set(ValidSiteKey, s)
	}
}

func GetSiteFromContext(ctx *gin.Context) (site.Site, bool) {
	val, ok := ctx.Get(ValidSiteKey)
	if !ok {
		return nil, false
	}
	var s site.Site
	s, ok = val.(site.Site)
	return s, ok
}

func ValidDeploymentMiddleware() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		s, ok := GetSiteFromContext(ctx)
		if !ok {
			// TODO: log
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		deploymentID := ctx.Param("deploymentID")
		d, err := s.GetDeployment(deploymentID)
		if err != nil {
			// TODO: log
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		ctx.Set(ValidDeploymentKey, d)
	}
}

func GetDeploymentFromContext(ctx *gin.Context) (deployment.Deployment, bool) {
	val, ok := ctx.Get(ValidSiteKey)
	if !ok {
		return nil, false
	}
	var d deployment.Deployment
	d, ok = val.(deployment.Deployment)
	return d, ok
}
