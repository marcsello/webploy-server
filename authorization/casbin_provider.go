package authorization

import (
	_ "embed"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/gin-gonic/gin"
	"net/http"
	"webploy-server/authentication"
)

type CasbinProvider struct {
	enforcer *casbin.Enforcer
}

//go:embed model.conf
var modelConfigString string

func NewCasbinProvider(policyFile string) (*CasbinProvider, error) {
	var err error

	var m model.Model
	m, err = model.NewModelFromString(modelConfigString)
	if err != nil {
		return nil, err
	}

	var e *casbin.Enforcer
	e, err = casbin.NewEnforcer(m, policyFile)
	if err != nil {
		return nil, err
	}

	return &CasbinProvider{
		enforcer: e,
	}, nil
}

func (cb *CasbinProvider) NewMiddleware(act string) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		user, ok := authentication.GetAuthenticatedUser(ctx)
		if !ok {
			//this should not happen
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		resource := ctx.Param("siteID")

		allowed, err := cb.enforcer.Enforce(user, resource, act)

		if err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			// TODO: log
			return
		}

		if !allowed {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		// All went fine
	}
}
