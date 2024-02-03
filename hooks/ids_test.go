package hooks

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"webploy-server/config"
)

func TestGetHookPathFromConfig(t *testing.T) {
	testConfig := config.HooksConfig{
		PreCreate:  "test1",
		PreFinish:  "test2",
		PostFinish: "test3",
		PreLive:    "test4",
		PostLive:   "test5",
	}

	testCases := []struct {
		name         string
		argHook      HookID
		argConfig    config.HooksConfig
		expectedPath string
		expectPanic  bool
	}{
		{
			name:         "happy__pre_create",
			argHook:      HookPreCreate,
			argConfig:    testConfig,
			expectedPath: testConfig.PreCreate,
		},
		{
			name:         "happy__pre_finish",
			argHook:      HookPreFinish,
			argConfig:    testConfig,
			expectedPath: testConfig.PreFinish,
		},
		{
			name:         "happy__post_finish",
			argHook:      HookPostFinish,
			argConfig:    testConfig,
			expectedPath: testConfig.PostFinish,
		},
		{
			name:         "happy__pre_live",
			argHook:      HookPreLive,
			argConfig:    testConfig,
			expectedPath: testConfig.PreLive,
		},
		{
			name:         "happy__post_live",
			argHook:      HookPostLive,
			argConfig:    testConfig,
			expectedPath: testConfig.PostLive,
		},
		{
			name:        "error__invalid",
			argHook:     HookID("asd"),
			argConfig:   testConfig,
			expectPanic: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			if tc.expectPanic {
				assert.Panics(t, func() {
					getHookPathFromConfig(tc.argHook, tc.argConfig)
				})
			} else {
				res := getHookPathFromConfig(tc.argHook, tc.argConfig)
				assert.Equal(t, tc.expectedPath, res)
			}

		})
	}
}
