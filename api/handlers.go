package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"webploy-server/authentication"
	"webploy-server/deployment"
	"webploy-server/deployment/info"
	"webploy-server/site"
)

func createDeployment(ctx *gin.Context) {
	user, ok := authentication.GetAuthenticatedUser(ctx)
	if !ok {
		// should not happen
		ctx.AbortWithStatus(http.StatusUnauthorized)
		// TODO: log
		return
	}

	var s site.Site
	s, ok = GetSiteFromContext(ctx)
	if !ok {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	var req NewDeploymentReq // TODO: limit meta size
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, ErrorResp{Err: err})
		// TODO: log
		return
	}

	// TODO: limit open deployment count

	var id string
	var d deployment.Deployment
	id, d, err = s.CreateNewDeployment(user, req.Meta)
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	var i info.DeploymentInfo
	i, err = d.GetFullInfo()
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
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

func uploadToDeployment(ctx *gin.Context) {
	user, ok := authentication.GetAuthenticatedUser(ctx)
	if !ok {
		// should not happen
		ctx.AbortWithStatus(http.StatusUnauthorized)
		// TODO: log
		return
	}

	var d deployment.Deployment
	_, d, ok = GetDeploymentFromContext(ctx)
	if !ok {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	// TODO: Limit uploads

	i, err := d.GetFullInfo()
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	if i.Creator != user {
		ctx.AbortWithStatus(http.StatusForbidden)
		// TODO: log
		return
	}

	err = d.AddFile(ctx, "TODO", ctx.Request.Body) // <- TODO: filename
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	ctx.Status(http.StatusCreated)
}

func finishDeployment(ctx *gin.Context) {
	user, ok := authentication.GetAuthenticatedUser(ctx)
	if !ok {
		// should not happen
		ctx.AbortWithStatus(http.StatusUnauthorized)
		// TODO: log
		return
	}

	var s site.Site
	s, ok = GetSiteFromContext(ctx)
	if !ok {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	var d deployment.Deployment
	var dID string
	dID, d, ok = GetDeploymentFromContext(ctx)
	if !ok {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	i, err := d.GetFullInfo()
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	if i.Creator != user {
		ctx.AbortWithStatus(http.StatusForbidden)
		// TODO: log
		return
	}

	// TODO run scripts

	err = d.Finish()
	if err != nil {
		if errors.Is(err, deployment.ErrDeploymentFinished) {
			// deployment already finished
			ctx.AbortWithStatus(http.StatusConflict)
			// TODO: log
			return
		}
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	i, err = d.GetFullInfo()
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	// set live on finish
	var setAsLive bool
	if s.GetConfig().GoLiveOnFinish {
		// TODO run scripts
		err = s.SetLiveDeploymentID(dID)
		if err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			// TODO: log
			return
		}
		setAsLive = true
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
	s, ok := GetSiteFromContext(ctx)
	if !ok {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	deployments, err := s.ListDeploymentIDs()
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	ctx.JSON(http.StatusOK, deployments)
}

func readDeployment(ctx *gin.Context) {
	s, ok := GetSiteFromContext(ctx)
	if !ok {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	var d deployment.Deployment
	var dID string
	dID, d, ok = GetDeploymentFromContext(ctx)
	if !ok {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	i, err := d.GetFullInfo()
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	var liveDID string
	liveDID, err = s.GetLiveDeploymentID()
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
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
	s, ok := GetSiteFromContext(ctx)
	if !ok {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	id, err := s.GetLiveDeploymentID()
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	var d deployment.Deployment
	d, err = s.GetDeployment(id)
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	var i info.DeploymentInfo
	i, err = d.GetFullInfo()
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
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
	s, ok := GetSiteFromContext(ctx)
	if !ok {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	var req LiveReq
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		// TODO: log
		return
	}

	err = s.SetLiveDeploymentID(req.ID)
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	var d deployment.Deployment
	d, err = s.GetDeployment(req.ID)
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	var i info.DeploymentInfo
	i, err = d.GetFullInfo()
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

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
