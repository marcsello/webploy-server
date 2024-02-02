package authorization

import (
	_ "embed"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"webploy-server/authentication"
)

type CasbinProvider struct {
	enforcer *casbin.Enforcer
	logger   *zap.Logger
}

//go:embed model.conf
var modelConfigString string

func NewCasbinProvider(policyFile string, logger *zap.Logger) (*CasbinProvider, error) {
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
		logger:   logger,
	}, nil
}

func (cb *CasbinProvider) NewMiddleware(act string) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		user, ok := authentication.GetAuthenticatedUser(ctx)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			cb.logger.Error("Failed to load user from context")
			return
		}

		resource := ctx.Param("siteID")

		l := cb.logger.With(zap.String("act", act), zap.String("resource", resource), zap.String("user", user))

		allowed, err := cb.enforcer.Enforce(user, resource, act)

		if err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			l.Error("Failed to check policy enforcement", zap.Error(err))
			return
		}

		if !allowed {
			ctx.AbortWithStatus(http.StatusForbidden)
			l.Debug("User unauthorized")
			return
		}

		// All went fine
		l.Debug("User authorized")
	}
}
