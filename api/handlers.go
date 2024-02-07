package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"os"
	"strings"
	"unicode/utf8"
	"webploy-server/adapters"
	"webploy-server/authentication"
	"webploy-server/authorization"
	"webploy-server/deployment"
	"webploy-server/deployment/info"
	"webploy-server/hooks"
	"webploy-server/site"
)

const MetaLengthLimit = 768 // unicode runes

func createDeployment(ctx *gin.Context) {
	l := GetLoggerFromContext(ctx)
	user, ok := authentication.GetAuthenticatedUser(ctx)
	if !ok {
		// should not happen
		ctx.Status(http.StatusInternalServerError)
		l.Error("Could not load user from context")
		return
	}

	s := GetSiteFromContext(ctx)

	// Parse request body
	var req NewDeploymentReq
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResp{Err: err})
		l.Warn("Could not un-marshal request body", zap.Error(err))
		return
	}

	metaRunesCount := utf8.RuneCountInString(req.Meta)
	if metaRunesCount > MetaLengthLimit { // measure in unicode runes
		ctx.JSON(http.StatusBadRequest, ErrorResp{ErrStr: "meta too long"})
		l.Warn("Meta is too long", zap.Int("metaRunesCount", metaRunesCount))
		return
	}

	// Limit open deployment count (0 = unlimited)
	if s.GetConfig().MaxOpen != 0 {
		var currentlyOpen uint
		err = s.IterDeployments(func(_ string, d deployment.Deployment, _ bool) (bool, error) {
			finished, e := d.IsFinished()
			if e != nil {
				return false, e
			}
			if !finished {
				currentlyOpen++
			}
			return true, nil // continue iteration
		})
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			l.Error("Error while iterating trough deployments", zap.Error(err))
			return
		}

		if currentlyOpen >= s.GetConfig().MaxOpen {
			err = fmt.Errorf("too many open deployments")
			ctx.JSON(http.StatusBadRequest, ErrorResp{Err: err})
			l.Warn("Could not create new deployment because deployment limit reached", zap.Error(err), zap.Uint("currentlyOpen", currentlyOpen), zap.Uint("MaxOpen", s.GetConfig().MaxOpen))
			return
		}

		l.Debug("Open deployment count check passed", zap.Uint("currentlyOpen", currentlyOpen), zap.Uint("MaxOpen", s.GetConfig().MaxOpen))
	} else {
		l.Debug("Skipping open deployment check because max deployments are unlimited.")
	}

	l.Debug("Executing hooks (if any)...")
	hookVars := hooks.HookVars{
		User:           user,
		DeploymentMeta: req.Meta,
	}
	err = hookVars.ReadFromSite(s)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		l.Error("Failed to load hook vars from site", zap.Error(err))
		return
	}
	ok, err = hooks.RunHook(ctx, s.GetConfig().Hooks, hooks.HookPreCreate, hookVars)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		l.Error("Failed to run hook", zap.Error(err))
		return
	}
	if !ok {
		ctx.JSON(http.StatusFailedDependency, ErrorResp{ErrStr: "prevented by hook"})
		l.Warn("Action is prevented by hook (non-zero exit code)") // since the hook logs are not connected to the handler logs in any way... this will be hard to debug...
		return
	}
	l.Debug("Hooks executed successfully")

	var id string
	var d deployment.Deployment
	id, d, err = s.CreateNewDeployment(user, req.Meta)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		l.Error("Failed to create new deployment", zap.Error(err))
		return
	}

	l.Info("New deployment created!", zap.String("deploymentID", id))

	var i info.DeploymentInfo
	i, err = d.GetFullInfo()
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		l.Error("Failed to read info for new deployment", zap.Error(err))
		return
	}

	resp := DeploymentInfoResp{
		Site:       s.GetName(),
		ID:         id,
		Creator:    user,
		CreatedAt:  i.CreatedAt,
		FinishedAt: nil,
		Meta:       i.Meta,
		IsLive:     false,
		IsFinished: false,
	}

	ctx.JSON(http.StatusCreated, resp)
}

func uploadFileToDeployment(ctx *gin.Context) {
	l := GetLoggerFromContext(ctx) // this panics
	user, ok := authentication.GetAuthenticatedUser(ctx)
	if !ok {
		// should not happen
		ctx.Status(http.StatusInternalServerError)
		l.Error("Could not load user from context")
		return
	}

	_, d := GetDeploymentFromContext(ctx)

	i, err := d.GetFullInfo()
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		l.Error("Failed to read info for deployment", zap.Error(err))
		return
	}

	var allowed bool
	allowed, err = ternaryEnforce(ctx, i.Creator == user, authorization.ActUploadSelf, authorization.ActUploadAny)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		l.Error("Failed to check for permission", zap.Error(err))
		return
	}
	if !allowed {
		ctx.JSON(http.StatusForbidden, ErrorResp{ErrStr: "no permission to upload into this"})
		l.Warn("Prevented upload to deployment, the user have no permission to do this", zap.String("deploymentCreator", i.Creator))
		return
	}

	// Load filename from header
	filename := ctx.GetHeader("X-Filename")
	if filename == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResp{ErrStr: "filename undefined"})
		l.Warn("X-Filename parameter is missing or empty")
		return
	}
	filename = strings.TrimLeft(filename, "/\\.") // this is just to silently fix common mistakes, relative paths are properly enforced on the deployment level
	l = l.With(zap.String("filename", filename))
	l.Debug("Target filename decoded")

	err = d.AddFile(ctx, filename, ctx.Request.Body) // <- Concurrent upload limiting handled here
	if err != nil {
		if errors.Is(err, deployment.ErrDeploymentFinished) {
			ctx.JSON(http.StatusBadRequest, ErrorResp{Err: err})
			l.Warn("Trying to upload to an already finished deployment", zap.Error(err))
			return
		}
		if errors.Is(err, deployment.ErrUploadInvalidPath) {
			ctx.JSON(http.StatusBadRequest, ErrorResp{Err: err})
			l.Warn("Trying to upload with an invalid path", zap.Error(err))
			return
		}
		if errors.Is(err, deployment.ErrTooManyConcurrentUploads) {
			ctx.JSON(http.StatusTooManyRequests, ErrorResp{Err: err})
			l.Warn("Too many pending uploads for deployment", zap.Error(err))
			return
		}
		if errors.Is(err, os.ErrExist) {
			ctx.JSON(http.StatusConflict, ErrorResp{Err: err})
			l.Warn("Trying to upload a file that already exists", zap.Error(err))
			return
		}

		ctx.Status(http.StatusInternalServerError)
		l.Error("Failed to upload file", zap.Error(err))
		return
	}

	l.Debug("New file uploaded!")
	ctx.Status(http.StatusCreated)
}

func uploadTarToDeployment(ctx *gin.Context) {
	l := GetLoggerFromContext(ctx) // this panics
	user, ok := authentication.GetAuthenticatedUser(ctx)
	if !ok {
		// should not happen
		ctx.Status(http.StatusInternalServerError)
		l.Error("Could not load user from context")
		return
	}

	_, d := GetDeploymentFromContext(ctx)

	i, err := d.GetFullInfo()
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		l.Error("Failed to read info for deployment", zap.Error(err))
		return
	}

	var allowed bool
	allowed, err = ternaryEnforce(ctx, i.Creator == user, authorization.ActUploadSelf, authorization.ActUploadAny)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		l.Error("Failed to check for permission", zap.Error(err))
		return
	}
	if !allowed {
		ctx.JSON(http.StatusForbidden, ErrorResp{ErrStr: "no permission to upload into this"})
		l.Warn("Prevented upload to deployment, the user have no permission to do this", zap.String("deploymentCreator", i.Creator))
		return
	}

	err = adapters.ExtractTarAdapter(ctx, l, d, ctx.Request.Body)
	if err != nil {
		if errors.Is(err, deployment.ErrDeploymentFinished) {
			ctx.JSON(http.StatusBadRequest, ErrorResp{Err: err})
			l.Warn("Trying to upload to an already finished deployment", zap.Error(err))
			return
		}
		if errors.Is(err, deployment.ErrUploadInvalidPath) {
			ctx.JSON(http.StatusBadRequest, ErrorResp{Err: err})
			l.Warn("Trying to upload with an invalid path", zap.Error(err))
			return
		}
		if errors.Is(err, deployment.ErrTooManyConcurrentUploads) {
			ctx.JSON(http.StatusTooManyRequests, ErrorResp{Err: err})
			l.Warn("Too many pending uploads for deployment", zap.Error(err))
			return
		}
		if errors.Is(err, os.ErrExist) {
			ctx.JSON(http.StatusConflict, ErrorResp{Err: err})
			l.Warn("Trying to upload a file that already exists", zap.Error(err))
			return
		}

		ctx.Status(http.StatusInternalServerError)
		l.Error("Failed to upload files", zap.Error(err))
		return
	}

	l.Debug("Files uploaded from tar archive!")
	ctx.Status(http.StatusCreated)

}

func finishDeployment(ctx *gin.Context) {
	l := GetLoggerFromContext(ctx) // this panics
	user, ok := authentication.GetAuthenticatedUser(ctx)
	if !ok {
		// should not happen
		ctx.Status(http.StatusUnauthorized)
		l.Error("Could not load user from context")
		return
	}

	s := GetSiteFromContext(ctx)
	dID, d := GetDeploymentFromContext(ctx)

	i, err := d.GetFullInfo()
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		l.Error("Failed to read info for deployment", zap.Error(err))
		return
	}

	var allowed bool
	allowed, err = ternaryEnforce(ctx, i.Creator == user, authorization.ActFinishSelf, authorization.ActFinishAny)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		l.Error("Failed to check for permission", zap.Error(err))
		return
	}
	if !allowed {
		ctx.JSON(http.StatusForbidden, ErrorResp{ErrStr: "no permission to finish this"})
		l.Warn("Prevented finishing the deployment, the user have no permission to finish this deployment", zap.String("deploymentCreator", i.Creator))
		return
	}

	l.Debug("Executing PreFinish hooks (if any)...")
	preFinishHookVars := hooks.HookVars{
		User:         user,
		DeploymentID: dID,
	}
	err = preFinishHookVars.ReadFromSiteAndDeployment(s, d)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		l.Error("Failed to read hook vars from site or deployment", zap.Error(err))
		return
	}

	ok, err = hooks.RunHook(ctx, s.GetConfig().Hooks, hooks.HookPreFinish, preFinishHookVars)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		l.Error("Failed to run hook", zap.Error(err))
		return
	}
	if !ok {
		ctx.JSON(http.StatusFailedDependency, ErrorResp{ErrStr: "finishing prevented by hook"})
		l.Warn("Action is prevented by hook (non-zero exit code)") // since the hook logs are not connected to the handler logs in any way... this will be hard to debug...
		return
	}
	l.Debug("Hooks executed successfully")

	err = d.Finish()
	if err != nil {
		if errors.Is(err, deployment.ErrDeploymentFinished) {
			// deployment already finished
			ctx.JSON(http.StatusConflict, ErrorResp{Err: err})
			l.Warn("Could not finish deployment because it is already finished", zap.Error(err))
			return
		}
		if errors.Is(err, deployment.ErrUploadPending) {
			// deployment already finished
			ctx.JSON(http.StatusConflict, ErrorResp{Err: err})
			l.Warn("Could not finish deployment because there is a pending upload", zap.Error(err))
			return
		}
		ctx.Status(http.StatusInternalServerError)
		l.Error("Could not finish deployment", zap.Error(err))
		return
	}

	l.Info("Finished deployment!", zap.Bool("GoLiveOnFinish", s.GetConfig().GoLiveOnFinish))

	l.Debug("Executing PostFinish hooks in the background (if any)...")
	postFinishHookVars := preFinishHookVars.Copy()
	go func() {
		_, err = hooks.RunHook(context.Background(), s.GetConfig().Hooks, hooks.HookPostFinish, postFinishHookVars)
		if err != nil {
			l.Error("Failed to run hook", zap.Error(err))
		}
		// stuff are logged by the hook runner as well
	}()

	i, err = d.GetFullInfo()
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		l.Error("Failed to read info for deployment (after finishing it)", zap.Error(err))
		return
	}

	// set live on finish
	var setAsLive bool
	if s.GetConfig().GoLiveOnFinish {

		l.Debug("Executing PreLive hooks (if any)...")
		preLiveHookVars := preFinishHookVars.Copy()
		ok, err = hooks.RunHook(ctx, s.GetConfig().Hooks, hooks.HookPreFinish, preLiveHookVars)
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			l.Error("Failed to run hook", zap.Error(err))
			return
		}
		if !ok {
			l.Warn("Setting as live is prevented by hook (non-zero exit code)") // since the hook logs are not connected to the handler logs in any way... this will be hard to debug...
		} else { // set as live only if the hook was successful
			err = s.SetLiveDeploymentID(dID)
			if err != nil {
				ctx.Status(http.StatusInternalServerError)
				l.Error("Failed to set deployment as live", zap.Error(err))
				return
			}
			setAsLive = true
			l.Info("Deployment set as live")
		}
	}

	if setAsLive {
		l.Debug("Executing PostLive hooks in the background (if any)...")
		postLiveHookVars := preFinishHookVars.Copy()
		postLiveHookVars.SiteCurrentLive = dID // this is the only field that should change
		go func() {
			_, err = hooks.RunHook(context.Background(), s.GetConfig().Hooks, hooks.HookPostLive, postLiveHookVars)
			if err != nil {
				l.Error("Failed to run hook", zap.Error(err))
			}
			// stuff are logged by the hook runner as well
		}()
	}

	// TODO: delete old

	resp := DeploymentInfoResp{
		Site:       s.GetName(),
		ID:         dID,
		Creator:    i.Creator,
		CreatedAt:  i.CreatedAt,
		FinishedAt: i.FinishedAt,
		Meta:       i.Meta,
		IsLive:     setAsLive,
		IsFinished: i.IsFinished(),
	}

	ctx.JSON(http.StatusOK, resp)
}

func listDeployments(ctx *gin.Context) {
	l := GetLoggerFromContext(ctx)
	s := GetSiteFromContext(ctx)

	deployments, err := s.ListDeploymentIDs()
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		l.Error("Failed to list deployments", zap.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, deployments)
}

func readDeployment(ctx *gin.Context) {
	l := GetLoggerFromContext(ctx)
	s := GetSiteFromContext(ctx)
	dID, d := GetDeploymentFromContext(ctx)

	i, err := d.GetFullInfo()
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		l.Error("Failed to read info for deployment", zap.Error(err))
		return
	}

	var liveDID string
	liveDID, err = s.GetLiveDeploymentID()
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		l.Error("Failed to read live deployment", zap.Error(err))
		return
	}

	resp := DeploymentInfoResp{
		Site:       s.GetName(),
		ID:         dID,
		Creator:    i.Creator,
		CreatedAt:  i.CreatedAt,
		FinishedAt: i.FinishedAt,
		Meta:       i.Meta,
		IsLive:     liveDID == dID,
		IsFinished: i.IsFinished(),
	}

	ctx.JSON(http.StatusOK, resp)
}

func readLiveDeployment(ctx *gin.Context) {
	l := GetLoggerFromContext(ctx)
	s := GetSiteFromContext(ctx)

	id, err := s.GetLiveDeploymentID()
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		l.Error("Failed to read live deployment", zap.Error(err))
		return
	}

	l = l.With(zap.String("deploymentID", id))

	var d deployment.Deployment
	d, err = s.GetDeployment(id)
	if err != nil {
		if errors.Is(err, site.ErrInvalidID) {
			ctx.JSON(http.StatusBadRequest, ErrorResp{Err: err})
			l.Warn("Tried to use an invalid deployment ID", zap.Error(err))
			return
		}

		ctx.Status(http.StatusInternalServerError)
		l.Error("Failed to read deployment", zap.Error(err))
		return
	}

	var i info.DeploymentInfo
	i, err = d.GetFullInfo()
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		l.Error("Failed to read info for deployment", zap.Error(err))
		return
	}

	resp := DeploymentInfoResp{
		Site:       s.GetName(),
		ID:         id,
		Creator:    i.Creator,
		CreatedAt:  i.CreatedAt,
		FinishedAt: i.FinishedAt,
		Meta:       i.Meta,
		IsLive:     true,
		IsFinished: i.IsFinished(),
	}

	ctx.JSON(http.StatusOK, resp)
}

func updateLiveDeployment(ctx *gin.Context) {
	l := GetLoggerFromContext(ctx) // this panics

	user, ok := authentication.GetAuthenticatedUser(ctx)
	if !ok {
		// should not happen
		ctx.Status(http.StatusUnauthorized)
		l.Error("Could not load user from context")
		return
	}

	s := GetSiteFromContext(ctx)

	// read request body
	var req LiveReq
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResp{Err: err})
		l.Warn("Could not un-marshal request body", zap.Error(err))
		return
	}

	l = l.With(zap.String("deploymentID", req.ID))

	// Load deployment (needed for hooks)
	var d deployment.Deployment
	d, err = s.GetDeployment(req.ID)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		l.Error("Failed to load deployment", zap.Error(err))
		return
	}

	// exec hooks
	l.Debug("Executing PreLive hooks (if any)...")
	preLiveHookVars := hooks.HookVars{
		User:         user,
		DeploymentID: req.ID,
	}
	err = preLiveHookVars.ReadFromSiteAndDeployment(s, d)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		l.Error("Failed to read hook vars from site or deployment", zap.Error(err))
		return
	}

	ok, err = hooks.RunHook(ctx, s.GetConfig().Hooks, hooks.HookPreFinish, preLiveHookVars)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		l.Error("Failed to run hook", zap.Error(err))
		return
	}
	if !ok {
		ctx.JSON(http.StatusFailedDependency, ErrorResp{ErrStr: "prevented by hook"})
		l.Warn("Action is prevented by hook (non-zero exit code)") // since the hook logs are not connected to the handler logs in any way... this will be hard to debug...
		return
	}
	l.Debug("Hooks executed successfully")

	// Actually set stuff as live
	err = s.SetLiveDeploymentID(req.ID)
	if err != nil {
		if errors.Is(err, site.ErrDeploymentNotExists) {
			ctx.JSON(http.StatusNotFound, ErrorResp{Err: err})
			l.Warn("Tried to set a missing deployment as live", zap.Error(err))
			return
		}
		if errors.Is(err, site.ErrInvalidID) {
			ctx.JSON(http.StatusBadRequest, ErrorResp{Err: err})
			l.Warn("Tried to use an invalid deployment ID", zap.Error(err))
			return
		}
		if errors.Is(err, site.ErrDeploymentNotFinished) {
			ctx.JSON(http.StatusBadRequest, ErrorResp{Err: err})
			l.Warn("Tried to set an unfinished deployment live", zap.Error(err))
			return
		}

		ctx.Status(http.StatusInternalServerError)
		l.Error("Could not update live deployment", zap.Error(err))
		return
	}

	// read back data, so we can return with it
	d, err = s.GetDeployment(req.ID)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		l.Error("Failed to read deployment (after setting it as live)", zap.Error(err))
		return
	}

	var i info.DeploymentInfo
	i, err = d.GetFullInfo()
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		l.Error("Failed to read info for deployment", zap.Error(err))
		return
	}

	l.Info("Live deployment updated")

	l.Debug("Executing PostLive hooks in the background (if any)...")
	postLiveHookVars := preLiveHookVars.Copy()
	postLiveHookVars.SiteCurrentLive = req.ID
	go func() {
		_, err = hooks.RunHook(context.Background(), s.GetConfig().Hooks, hooks.HookPostLive, postLiveHookVars)
		if err != nil {
			l.Error("Failed to run hook", zap.Error(err))
		}
		// stuff are logged by the hook runner as well
	}()

	resp := DeploymentInfoResp{
		Site:       s.GetName(),
		ID:         req.ID,
		Creator:    i.Creator,
		CreatedAt:  i.CreatedAt,
		FinishedAt: i.FinishedAt,
		Meta:       i.Meta,
		IsLive:     true,
		IsFinished: i.IsFinished(),
	}

	ctx.JSON(http.StatusOK, resp)
}
