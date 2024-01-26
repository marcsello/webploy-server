package info

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDeploymentInfo_Simple(t *testing.T) {
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

func TestDeploymentInfo_CopyDiffPtr(t *testing.T) {
	now := time.Now()
	info1 := DeploymentInfo{
		Creator:        "test",
		CreatedAt:      time.Time{},
		State:          DeploymentStateOpen,
		FinishedAt:     &now,
		LastActivityAt: time.Time{},
	}
	info2 := info1.Copy()

	assert.False(t, info1.FinishedAt == info2.FinishedAt)
}

func TestDeploymentInfo_Copy(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name string
		info DeploymentInfo
	}{
		{
			name: "simple",
			info: DeploymentInfo{
				Creator:        "test",
				CreatedAt:      time.Time{},
				State:          DeploymentStateOpen,
				FinishedAt:     nil,
				LastActivityAt: time.Time{},
			},
		},
		{
			name: "simple_2",
			info: DeploymentInfo{
				Creator:        "test",
				CreatedAt:      now,
				State:          DeploymentStateOpen,
				FinishedAt:     &now,
				LastActivityAt: now,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cpy := tc.info.Copy()
			assert.True(t, tc.info.Equals(cpy))
			assert.Equal(t, tc.info, cpy)
		})
	}

}

func TestDeploymentInfo_Equal(t *testing.T) {
	now := time.Now()
	d1 := time.Date(2002, 01, 12, 12, 0, 0, 0, time.UTC)
	d11 := time.Date(2002, 01, 12, 12, 0, 0, 0, time.UTC)
	d2 := time.Date(2002, 01, 12, 12, 0, 2, 0, time.UTC)
	testCases := []struct {
		name          string
		A             DeploymentInfo
		B             DeploymentInfo
		expectedEqual bool
	}{
		{
			name:          "happy__eq_empty",
			expectedEqual: true,
		},
		{
			name: "happy__eq_simple_noptr",
			A: DeploymentInfo{
				Creator:        "test",
				CreatedAt:      now,
				State:          DeploymentStateOpen,
				FinishedAt:     nil,
				LastActivityAt: now,
			},
			B: DeploymentInfo{
				Creator:        "test",
				CreatedAt:      now,
				State:          DeploymentStateOpen,
				FinishedAt:     nil,
				LastActivityAt: now,
			},
			expectedEqual: true,
		},
		{
			name: "happy__eq_simple_with_same_ptr",
			A: DeploymentInfo{
				Creator:        "test",
				CreatedAt:      now,
				State:          DeploymentStateOpen,
				FinishedAt:     &now,
				LastActivityAt: now,
			},
			B: DeploymentInfo{
				Creator:        "test",
				CreatedAt:      now,
				State:          DeploymentStateOpen,
				FinishedAt:     &now,
				LastActivityAt: now,
			},
			expectedEqual: true,
		},
		{
			name: "happy__eq_simple_with_diff_ptr",
			A: DeploymentInfo{
				Creator:        "test",
				CreatedAt:      now,
				State:          DeploymentStateOpen,
				FinishedAt:     &d1,
				LastActivityAt: now,
			},
			B: DeploymentInfo{
				Creator:        "test",
				CreatedAt:      now,
				State:          DeploymentStateOpen,
				FinishedAt:     &d11,
				LastActivityAt: now,
			},
			expectedEqual: true,
		},
		{
			name: "happy__neq_simple",
			A: DeploymentInfo{
				Creator:        "test",
				CreatedAt:      now,
				State:          DeploymentStateOpen,
				FinishedAt:     nil,
				LastActivityAt: now,
			},
			B: DeploymentInfo{
				Creator:        "test2",
				CreatedAt:      now,
				State:          DeploymentStateOpen,
				FinishedAt:     nil,
				LastActivityAt: now,
			},
			expectedEqual: false,
		},
		{
			name: "happy__neq_simple_2",
			A: DeploymentInfo{
				Creator:        "test",
				CreatedAt:      d1,
				State:          DeploymentStateOpen,
				FinishedAt:     nil,
				LastActivityAt: now,
			},
			B: DeploymentInfo{
				Creator:        "test",
				CreatedAt:      d2,
				State:          DeploymentStateOpen,
				FinishedAt:     nil,
				LastActivityAt: now,
			},
			expectedEqual: false,
		},
		{
			name: "happy__neq_simple_3",
			A: DeploymentInfo{
				Creator:        "test",
				CreatedAt:      d1,
				State:          DeploymentStateOpen,
				FinishedAt:     nil,
				LastActivityAt: d1,
			},
			B: DeploymentInfo{
				Creator:        "test",
				CreatedAt:      d1,
				State:          DeploymentStateOpen,
				FinishedAt:     nil,
				LastActivityAt: d2,
			},
			expectedEqual: false,
		},
		{
			name: "happy__neq_simple_4",
			A: DeploymentInfo{
				Creator:        "test",
				CreatedAt:      d1,
				State:          DeploymentStateOpen,
				FinishedAt:     nil,
				LastActivityAt: d1,
			},
			B: DeploymentInfo{
				Creator:        "test",
				CreatedAt:      d1,
				State:          DeploymentStateFinished,
				FinishedAt:     nil,
				LastActivityAt: d1,
			},
			expectedEqual: false,
		},
		{
			name: "happy__neq_ptr",
			A: DeploymentInfo{
				Creator:        "test",
				CreatedAt:      now,
				State:          DeploymentStateOpen,
				FinishedAt:     &now,
				LastActivityAt: now,
			},
			B: DeploymentInfo{
				Creator:        "test",
				CreatedAt:      now,
				State:          DeploymentStateOpen,
				FinishedAt:     nil,
				LastActivityAt: now,
			},
			expectedEqual: false,
		},
		{
			name: "happy__neq_ptr_2",
			A: DeploymentInfo{
				Creator:        "test",
				CreatedAt:      now,
				State:          DeploymentStateOpen,
				FinishedAt:     &d1,
				LastActivityAt: now,
			},
			B: DeploymentInfo{
				Creator:        "test",
				CreatedAt:      now,
				State:          DeploymentStateOpen,
				FinishedAt:     &d2,
				LastActivityAt: now,
			},
			expectedEqual: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedEqual, tc.A.Equals(tc.B))
			assert.Equal(t, tc.expectedEqual, tc.B.Equals(tc.A))
			if tc.expectedEqual {
				assert.Equal(t, tc.A, tc.B)
			} else {
				assert.NotEqual(t, tc.A, tc.B)
			}
		})
	}
}

func TestDeploymentInfo_EqualSelfAfterJSONMarshal(t *testing.T) {
	now := time.Now() // <- this is evil
	d1 := time.Date(2002, 01, 12, 12, 0, 0, 0, time.UTC)
	testCases := []struct {
		name string
		info DeploymentInfo
	}{
		{
			name: "test_simple",
			info: DeploymentInfo{
				Creator:        "test",
				CreatedAt:      now,
				State:          DeploymentStateOpen,
				FinishedAt:     nil,
				LastActivityAt: now,
			},
		},
		{
			name: "test_with_ptr",
			info: DeploymentInfo{
				Creator:        "test",
				CreatedAt:      now,
				State:          DeploymentStateOpen,
				FinishedAt:     &now,
				LastActivityAt: now,
			},
		},
		{
			name: "test_with_other_ptr",
			info: DeploymentInfo{
				Creator:        "test",
				CreatedAt:      now,
				State:          DeploymentStateOpen,
				FinishedAt:     &d1,
				LastActivityAt: now,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			var b []byte
			b, _ = json.Marshal(tc.info)

			var unmarshaledInfo DeploymentInfo
			_ = json.Unmarshal(b, &unmarshaledInfo)

			assert.True(t, tc.info.Equals(unmarshaledInfo))
			assert.True(t, unmarshaledInfo.Equals(tc.info))

		})
	}
}

func TestDeploymentInfo_IsFinished(t *testing.T) {

	info1 := DeploymentInfo{
		State: DeploymentStateOpen,
	}
	assert.False(t, info1.IsFinished())

	info2 := DeploymentInfo{
		State: DeploymentStateFinished,
	}
	assert.True(t, info2.IsFinished())

}
