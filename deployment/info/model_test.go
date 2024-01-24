package info

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDeploymentInfo_Copy(t *testing.T) {
	info1 := DeploymentInfo{
		Creator:        "test",
		CreatedAt:      time.Time{},
		State:          DeploymentStateOpen,
		FinishedAt:     nil,
		LastActivityAt: time.Time{},
	}
	info2 := info1.Copy()

	assert.True(t, info1.Equals(info2))
}
