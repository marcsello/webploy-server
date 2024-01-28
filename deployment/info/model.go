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
	Meta           string          `json:"meta"` // provided by the creator on creation
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
		Meta:           i.Meta,
	}
	if i.FinishedAt != nil {
		val := *i.FinishedAt
		cpy.FinishedAt = &val
	}
	return cpy
}

func (i *DeploymentInfo) Equals(o DeploymentInfo) bool {

	// we use UnixNano for time comparison, because otherwise JSON marshaled-unmarshaled values would fail, since go drops the monotonic part
	// (which is fair, but still needs a workaround)

	// highly magic check by value
	if i.FinishedAt != o.FinishedAt {
		if i.FinishedAt != nil && o.FinishedAt != nil {
			if i.FinishedAt.UnixNano() != o.FinishedAt.UnixNano() {
				return false
			}
		} else {
			return false
		}
	}

	return i.Creator == o.Creator &&
		i.CreatedAt.UnixNano() == o.CreatedAt.UnixNano() &&
		i.State == o.State &&
		i.LastActivityAt.UnixNano() == o.LastActivityAt.UnixNano() &&
		i.Meta == o.Meta
}
