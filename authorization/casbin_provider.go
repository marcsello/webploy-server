package authorization

import (
	_ "embed"
	"fmt"
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	"github.com/gin-gonic/gin"
	"github.com/marcsello/webploy-server/authentication"
	"go.uber.org/zap"
	"net/http"
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
		logger.Error("Failed to build the casbin model", zap.Error(err))
		return nil, err
	}

	adapter := fileadapter.NewAdapter(policyFile)

	var e *casbin.Enforcer
	e, err = casbin.NewEnforcer(m, adapter)
	if err != nil {
		logger.Error("Failed to initialize casbin enforcer", zap.Error(err))
		return nil, err
	}

	return &CasbinProvider{
		enforcer: e,
		logger:   logger,
	}, nil
}

const AuthZEnforcerFuncKey = "authz_enforcer_func"

type EnforcerFunction func(string) (bool, error)

func (cb *CasbinProvider) NewMiddleware(acts ...string) gin.HandlerFunc {

	return func(ctx *gin.Context) {
		user, ok := authentication.GetAuthenticatedUser(ctx)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			cb.logger.Error("Failed to load user from context")
			return
		}

		resource := ctx.Param("siteName") // read the url param directly

		l := cb.logger.With(zap.Strings("acts", acts), zap.String("resource", resource), zap.String("user", user))

		enforcerFunc := EnforcerFunction(func(act string) (bool, error) {
			return cb.enforcer.Enforce(user, resource, act)
		})

		for _, act := range acts { // enforce all acts one-by-one
			allowed, err := enforcerFunc(act)

			if err != nil {
				ctx.AbortWithStatus(http.StatusInternalServerError)
				l.Error("Failed to check policy enforcement", zap.String("act", act), zap.Error(err))
				return
			}

			if !allowed {
				ctx.AbortWithStatus(http.StatusForbidden)
				l.Debug("User unauthorized", zap.String("act", act))
				return
			}

			l.Debug("Act check passed", zap.String("act", act))
		}

		ctx.Set(AuthZEnforcerFuncKey, enforcerFunc)

		// All went fine
		l.Debug("Authorization completed")
	}
}

func EnforceAuthZ(ctx *gin.Context, act string) (bool, error) {
	val, ok := ctx.Get(AuthZEnforcerFuncKey)
	if !ok {
		return false, fmt.Errorf("could not load authz enforcer from context")
	}
	var fn EnforcerFunction
	fn, ok = val.(EnforcerFunction)
	if !ok {
		return false, fmt.Errorf("could not cast authz enforcer to type")
	}
	return fn(act)
}
