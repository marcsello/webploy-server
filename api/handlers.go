package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"webploy-server/authentication"
)

func createDeployment(ctx *gin.Context) {
	s, ok := GetSiteFromContext(ctx)
	if !ok {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	var user string
	user, ok = authentication.GetAuthenticatedUser(ctx)
	if !ok {
		// should not happen
		ctx.AbortWithStatus(http.StatusUnauthorized)
		// TODO: log
		return
	}

	deployment, err := s.CreateNewDeployment(user)
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"id": deployment.ID()})
}

func uploadToDeployment(ctx *gin.Context) {
	// TODO
}

func finishDeployment(ctx *gin.Context) {
	// TODO
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

func readCurrentDeployment(ctx *gin.Context) {
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

	ctx.JSON(http.StatusOK, gin.H{"id": id})
}

func updateCurrentDeployment(ctx *gin.Context) {
	// TODO
}
