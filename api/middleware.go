package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"webploy-server/site"
)

const ValidSiteKey = "site"

func ValidSiteMiddleware() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		siteName := ctx.Param("siteName")
		s, ok := site.GetSite(siteName)
		if !ok {
			ctx.AbortWithStatus(http.StatusNotFound)
			return
		}

		ctx.Set(ValidSiteKey, s)
	}
}

func GetSiteFromContext(ctx *gin.Context) (*site.Site, bool) {
	val, ok := ctx.Get(ValidSiteKey)
	if !ok {
		return nil, false
	}
	var s *site.Site
	s, ok = val.(*site.Site)
	return s, ok
}
