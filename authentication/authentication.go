package authentication

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/marcsello/webploy-server/config"
)

func InitAuthenticator(cfg config.AuthenticationProviderConfig) (Provider, error) {

	// For now we support ONLY ONE auth provider to be configured
	// Otherwise we would have to deal with realms

	if cfg.BasicAuth != nil {
		// load basic auth module
		return NewBasicAuthProvider(cfg.BasicAuth.HTPasswdFile)
	} else {
		return nil, fmt.Errorf("authentcation method not defined")
	}

}

func GetAuthenticatedUser(ctx *gin.Context) (string, bool) {
	bu, ok := ctx.Get(ContextAuthenticatedUserKey)
	if !ok {
		return "", false
	}
	var username string
	username, ok = bu.(string)
	if !ok {
		return "", false
	}
	return username, true
}
