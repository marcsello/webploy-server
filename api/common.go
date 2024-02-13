package api

import (
	"github.com/gin-gonic/gin"
	"github.com/marcsello/webploy-server/authorization"
	"go.uber.org/zap"
)

func ternaryEnforce(ctx *gin.Context, isSelf bool, actSelf, actAny string) (bool, error) {
	l := GetLoggerFromContext(ctx).With(zap.Bool("isSelf", isSelf), zap.String("actSelf", actSelf), zap.String("actAny", actAny))

	var allowed bool
	var err error

	// If this deployment is created by us, we first check if we allowed to finish our own deployment
	if isSelf {
		allowed, err = authorization.EnforceAuthZ(ctx, actSelf)
		if err != nil {
			l.Error("Failed to check for self act", zap.Error(err))
			return false, err
		}
	}

	// if it was either not created by us, or we weren't allowed to finish our own, then check if we allowed to finish any
	if !allowed {
		allowed, err = authorization.EnforceAuthZ(ctx, actAny)
		if err != nil {
			l.Error("Failed to check for any act", zap.Error(err))
			return false, err
		}
	}

	l.Debug("Ternary access check completed", zap.Bool("allowed", allowed))
	return allowed, nil
}
