package info

import "time"

type DeploymentState string

const (
	DeploymentStateOpen     DeploymentState = "open"
	DeploymentStateFinished DeploymentState = "finished"
)

type DeploymentInfo struct {
	Creator        string          `json:"creator"`
	CreatedAt      time.Time       `json:"created_at"`
	State          DeploymentState `json:"state"`
	FinishedAt     *time.Time      `json:"finished_at"`
	LastActivityAt time.Time       `json:"last_activity_at"`
}

func (i *DeploymentInfo) IsFinished() bool {
	return i.State == DeploymentStateFinished
}
