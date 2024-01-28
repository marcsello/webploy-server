package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"webploy-server/authentication"
	"webploy-server/deployment"
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

	id, _, err := s.CreateNewDeployment(user)
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"id": id})
}

func uploadToDeployment(ctx *gin.Context) {
	d, ok := GetDeploymentFromContext(ctx)
	if !ok {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	err := d.AddFile(ctx, "TODO", ctx.Request.Body) // <- TODO: filename
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	ctx.Status(http.StatusCreated)
}

func finishDeployment(ctx *gin.Context) {
	d, ok := GetDeploymentFromContext(ctx)
	if !ok {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	// TODO: enforce same-user ???

	err := d.Finish()
	if err != nil {
		if errors.Is(err, deployment.ErrDeploymentFinished) {
			// deployment already finished
			ctx.AbortWithStatus(http.StatusBadRequest)
			// TODO: log
			return
		}
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	ctx.Status(http.StatusOK)
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
	s, ok := GetSiteFromContext(ctx)
	if !ok {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	err := s.SetLiveDeploymentID("TODO") // <- TODO
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		// TODO: log
		return
	}

	ctx.Status(http.StatusOK)
}
