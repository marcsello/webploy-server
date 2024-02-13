package deployment

import (
	"fmt"
	"github.com/marcsello/webploy-server/deployment/info"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDeploymentImpl_IsFinished(t *testing.T) {
	testCases := []struct {
		name     string
		state    info.DeploymentState
		finished bool
		error    error
	}{
		{
			name:     "happy__simple_true",
			state:    info.DeploymentStateFinished,
			finished: true,
		},
		{
			name:     "happy__simple_false",
			state:    info.DeploymentStateOpen,
			finished: false,
		},
		{
			name:  "error__any",
			error: fmt.Errorf("test error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := DeploymentImpl{
				infoProvider: &info.InfoProviderMock{
					TxFn: func(readonly bool, txFunc info.InfoTransaction) error {
						assert.True(t, readonly)

						if tc.error != nil {
							return tc.error
						}

						i := &info.DeploymentInfo{
							State: tc.state,
						}
						return txFunc(i)
					},
				},
			}

			finished, err := d.IsFinished()
			if tc.error != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.error.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.finished, finished)
			}
		})
	}

}

func TestDeploymentImpl_Creator(t *testing.T) {
	testCases := []struct {
		name    string
		state   info.DeploymentState
		creator string
		error   error
	}{
		{
			name:    "happy__simple",
			state:   info.DeploymentStateFinished,
			creator: "test1",
		},
		{
			name:  "error__any",
			error: fmt.Errorf("test error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := DeploymentImpl{
				infoProvider: &info.InfoProviderMock{
					TxFn: func(readonly bool, txFunc info.InfoTransaction) error {
						assert.True(t, readonly)

						if tc.error != nil {
							return tc.error
						}

						i := &info.DeploymentInfo{
							Creator: tc.creator,
						}
						return txFunc(i)
					},
				},
			}

			creator, err := d.Creator()
			if tc.error != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.error.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.creator, creator)
			}
		})
	}

}

func TestDeploymentImpl_LastActivity(t *testing.T) {
	testCases := []struct {
		name         string
		lastActivity time.Time
		error        error
	}{
		{
			name:         "happy__simple",
			lastActivity: time.Now(),
		},
		{
			name:  "error__any",
			error: fmt.Errorf("test error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := DeploymentImpl{
				infoProvider: &info.InfoProviderMock{
					TxFn: func(readonly bool, txFunc info.InfoTransaction) error {
						assert.True(t, readonly)

						if tc.error != nil {
							return tc.error
						}

						i := &info.DeploymentInfo{
							LastActivityAt: tc.lastActivity,
						}
						return txFunc(i)
					},
				},
			}

			ts, err := d.LastActivity()
			if tc.error != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.error.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.lastActivity, ts)
			}
		})
	}

}

func TestDeploymentImpl_GetFullInfo(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name  string
		info  info.DeploymentInfo
		error error
	}{
		{
			name: "happy__simple_1",
			info: info.DeploymentInfo{},
		},
		{
			name: "happy__simple_2",
			info: info.DeploymentInfo{
				Creator:        "test",
				CreatedAt:      now,
				State:          info.DeploymentStateOpen,
				FinishedAt:     nil,
				LastActivityAt: now,
				Meta:           "test",
			},
		},
		{
			name: "happy__simple_3",
			info: info.DeploymentInfo{
				Creator:        "test",
				CreatedAt:      now,
				State:          info.DeploymentStateOpen,
				FinishedAt:     &now,
				LastActivityAt: now,
				Meta:           "test",
			},
		},
		{
			name:  "error__any",
			error: fmt.Errorf("test error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := DeploymentImpl{
				infoProvider: &info.InfoProviderMock{
					TxFn: func(readonly bool, txFunc info.InfoTransaction) error {
						assert.True(t, readonly)

						if tc.error != nil {
							return tc.error
						}

						return txFunc(&tc.info)
					},
				},
			}

			i, err := d.GetFullInfo()
			if tc.error != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.error.Error())
			} else {
				assert.NoError(t, err)
				assert.True(t, tc.info.Equals(i))
				assert.True(t, i.Equals(tc.info))
			}
		})
	}

}
