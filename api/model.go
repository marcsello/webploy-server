package api

import (
	"encoding/json"
	"time"
)

// NewDeploymentReq is expected to be sent by the user when creating a new deployment
type NewDeploymentReq struct {
	Meta string `json:"meta,omitempty"`
}

// DeploymentInfoResp is sent after creating a new deployment, requesting info, finishing, querying live etc.
type DeploymentInfoResp struct {
	Site       string     `json:"site"`
	ID         string     `json:"id"`
	Creator    string     `json:"creator"`
	CreatedAt  time.Time  `json:"created_at"`
	FinishedAt *time.Time `json:"finished_at"`
	Meta       string     `json:"meta,omitempty"`
	IsLive     bool       `json:"is_live"`
	IsFinished bool       `json:"is_finished"`
}

// LiveReq is provided by the user when updating the ID of the live deployment
type LiveReq struct {
	ID string `json:"id"`
}

// ErrorResp sent on any error happened
type ErrorResp struct {
	Err error
}

func (r ErrorResp) MarshalJSON() ([]byte, error) {
	var errorMsg string
	if r.Err != nil {
		errorMsg = r.Err.Error()
	}
	return json.Marshal(struct {
		Error string `json:"err"`
	}{
		Error: errorMsg,
	})
}
