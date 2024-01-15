package deployment

import "time"

type State struct {
	Creator        string    `json:"creator"`
	CreatedAt      time.Time `json:"created_at"`
	IsFinished     bool      `json:"is_finished"`
	FinishedAt     time.Time `json:"finished_at"`
	LastActivityAt time.Time `json:"last_activity_at"`
}
