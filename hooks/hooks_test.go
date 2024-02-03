package hooks

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
	"testing"
	"webploy-server/config"
)

func TestRunHook(t *testing.T) {
	testErr := fmt.Errorf("test error")

	testCases := []struct {
		name string

		argHooksConfig config.HooksConfig
		argHook        HookID
		argVars        HookVars

		expectedOk  bool
		expectedErr error

		expectExcCalled     bool
		excExpectedName     string
		excExpectedArgs     []string
		excExpectedExtraEnv []string
		excRetExitCode      int
		excRetOutput        []byte
		excRetError         error
	}{
		{
			name: "happy__simple",

			argHooksConfig: config.HooksConfig{
				PostFinish: "test_path",
			},
			argHook: HookPostFinish,
			argVars: HookVars{
				User:              "test1",
				SiteName:          "test2",
				SitePath:          "test3",
				SiteCurrentLive:   "test4",
				DeploymentID:      "test5",
				DeploymentCreator: "test6",
				DeploymentMeta:    "test7",
				DeploymentPath:    "test8",
			},
			expectedOk:  true,
			expectedErr: nil,

			expectExcCalled: true,
			excExpectedName: "test_path",
			excExpectedArgs: []string{string(HookPostFinish), "test8"},
			excExpectedExtraEnv: []string{
				"WEBPLOY_USER=test1",
				"WEBPLOY_SITE=test2",
				"WEBPLOY_SITE_PATH=test3",
				"WEBPLOY_SITE_CURRENT_LIVE=test4",
				"WEBPLOY_DEPLOYMENT_ID=test5",
				"WEBPLOY_DEPLOYMENT_CREATOR=test6",
				"WEBPLOY_DEPLOYMENT_META=test7",
				"WEBPLOY_DEPLOYMENT_PATH=test8",
				"WEBPLOY_HOOK=" + string(HookPostFinish),
			},
			excRetExitCode: 0,
			excRetOutput:   []byte("hello world"),
			excRetError:    nil,
		},
		{
			name: "happy__simple2",

			argHooksConfig: config.HooksConfig{
				PreCreate: "test_path",
			},
			argHook: HookPreCreate,
			argVars: HookVars{
				User:              "test1",
				SiteName:          "test2",
				SitePath:          "test3",
				SiteCurrentLive:   "test4",
				DeploymentID:      "",
				DeploymentCreator: "",
				DeploymentMeta:    "test7",
				DeploymentPath:    "",
			},
			expectedOk:  true,
			expectedErr: nil,

			expectExcCalled: true,
			excExpectedName: "test_path",
			excExpectedArgs: []string{string(HookPreCreate)},
			excExpectedExtraEnv: []string{
				"WEBPLOY_USER=test1",
				"WEBPLOY_SITE=test2",
				"WEBPLOY_SITE_PATH=test3",
				"WEBPLOY_SITE_CURRENT_LIVE=test4",
				"WEBPLOY_DEPLOYMENT_ID=",
				"WEBPLOY_DEPLOYMENT_CREATOR=",
				"WEBPLOY_DEPLOYMENT_META=test7",
				"WEBPLOY_DEPLOYMENT_PATH=",
				"WEBPLOY_HOOK=" + string(HookPreCreate),
			},
			excRetExitCode: 0,
			excRetOutput:   []byte("hello world"),
			excRetError:    nil,
		},
		{
			name: "happy__non_zero",

			argHooksConfig: config.HooksConfig{
				PreCreate: "test_path",
			},
			argHook: HookPreCreate,
			argVars: HookVars{
				User:              "test1",
				SiteName:          "test2",
				SitePath:          "test3",
				SiteCurrentLive:   "test4",
				DeploymentID:      "",
				DeploymentCreator: "",
				DeploymentMeta:    "test7",
				DeploymentPath:    "",
			},
			expectedOk:  false,
			expectedErr: nil,

			expectExcCalled: true,
			excExpectedName: "test_path",
			excExpectedArgs: []string{string(HookPreCreate)},
			excExpectedExtraEnv: []string{
				"WEBPLOY_USER=test1",
				"WEBPLOY_SITE=test2",
				"WEBPLOY_SITE_PATH=test3",
				"WEBPLOY_SITE_CURRENT_LIVE=test4",
				"WEBPLOY_DEPLOYMENT_ID=",
				"WEBPLOY_DEPLOYMENT_CREATOR=",
				"WEBPLOY_DEPLOYMENT_META=test7",
				"WEBPLOY_DEPLOYMENT_PATH=",
				"WEBPLOY_HOOK=" + string(HookPreCreate),
			},
			excRetExitCode: 1,
			excRetOutput:   []byte("hello world"),
			excRetError:    nil,
		},
		{
			name: "happy__no_hook_configured",

			argHooksConfig: config.HooksConfig{
				PreCreate: "",
			},
			argHook: HookPreCreate,
			argVars: HookVars{
				User:              "test1",
				SiteName:          "test2",
				SitePath:          "test3",
				SiteCurrentLive:   "test4",
				DeploymentID:      "",
				DeploymentCreator: "",
				DeploymentMeta:    "test7",
				DeploymentPath:    "",
			},
			expectedOk:  true,
			expectedErr: nil,

			expectExcCalled: false,
		},
		{
			name: "error__exec_fail",

			argHooksConfig: config.HooksConfig{
				PreCreate: "test_path",
			},
			argHook: HookPreCreate,
			argVars: HookVars{
				User:              "test1",
				SiteName:          "test2",
				SitePath:          "test3",
				SiteCurrentLive:   "test4",
				DeploymentID:      "",
				DeploymentCreator: "",
				DeploymentMeta:    "test7",
				DeploymentPath:    "",
			},
			expectedOk:  false,
			expectedErr: testErr,

			expectExcCalled: true,
			excExpectedName: "test_path",
			excExpectedArgs: []string{string(HookPreCreate)},
			excExpectedExtraEnv: []string{
				"WEBPLOY_USER=test1",
				"WEBPLOY_SITE=test2",
				"WEBPLOY_SITE_PATH=test3",
				"WEBPLOY_SITE_CURRENT_LIVE=test4",
				"WEBPLOY_DEPLOYMENT_ID=",
				"WEBPLOY_DEPLOYMENT_CREATOR=",
				"WEBPLOY_DEPLOYMENT_META=test7",
				"WEBPLOY_DEPLOYMENT_PATH=",
				"WEBPLOY_HOOK=" + string(HookPreCreate),
			},
			excRetExitCode: 0,
			excRetOutput:   nil,
			excRetError:    testErr,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger = zaptest.NewLogger(t)
			var excCalled bool
			exc = func(_ context.Context, name string, args []string, extraEnv []string) (int, []byte, error) {
				// Mock "exec"
				excCalled = true
				assert.Equal(t, tc.excExpectedName, name)
				assert.Equal(t, tc.excExpectedArgs, args)
				assert.ElementsMatch(t, tc.excExpectedExtraEnv, extraEnv)
				return tc.excRetExitCode, tc.excRetOutput, tc.excRetError
			}

			retOk, retErr := RunHook(context.Background(), tc.argHooksConfig, tc.argHook, tc.argVars)

			assert.Equal(t, tc.expectExcCalled, excCalled)
			assert.Equal(t, tc.expectedOk, retOk)

			if tc.expectedErr != nil {
				assert.Error(t, retErr)
				assert.Contains(t, retErr.Error(), tc.expectedErr.Error())
			} else {
				assert.NoError(t, retErr)
			}

		})
	}

}
