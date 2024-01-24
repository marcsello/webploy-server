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

func (i *DeploymentInfo) Copy() DeploymentInfo {
	cpy := DeploymentInfo{
		Creator:        i.Creator,
		CreatedAt:      i.CreatedAt,
		State:          i.State,
		FinishedAt:     nil,
		LastActivityAt: i.LastActivityAt,
	}
	if i.FinishedAt != nil {
		cpy.FinishedAt = &*i.FinishedAt
	}
	return cpy
}

func (i *DeploymentInfo) Equals(o DeploymentInfo) bool {

	// highly magic check by value
	if i.FinishedAt != o.FinishedAt {
		if i.FinishedAt != nil && o.FinishedAt != nil {
			if *i.FinishedAt != *o.FinishedAt {
				return false
			}
		} else {
			return false
		}
	}

	return i.Creator == o.Creator &&
		i.CreatedAt == i.CreatedAt &&
		i.State == o.State &&
		i.LastActivityAt == o.LastActivityAt
}
