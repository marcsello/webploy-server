package config

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
	"os"
	"path"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	testCases := []struct {
		name           string
		configYAML     string
		dontCreateFile bool
		expectedConfig WebployConfig
		expectedErr    error
	}{
		{
			name:       "happy__defaults",
			configYAML: `---`,
			expectedConfig: WebployConfig{
				Listen: ListenConfig{
					BindAddr:  ":8000",
					EnableTLS: false,
				},
				Sites: SitesConfig{
					Root:  "/var/www",
					Sites: []SiteConfig{},
				},
				Authorization: AuthorizationProviderConfig{
					PolicyFile: "/etc/webploy/policy.csv",
				},
			},
		},
		{
			name: "happy__defaults_extended",
			configYAML: `---
authentication:
  basic_auth: {}

sites:
  sites:
    - name: test1
`,
			expectedConfig: WebployConfig{
				Listen: ListenConfig{
					BindAddr:  ":8000",
					EnableTLS: false,
				},
				Sites: SitesConfig{
					Root: "/var/www",
					Sites: []SiteConfig{
						{
							Name:                 "test1",
							MaxHistory:           2,
							MaxOpen:              2,
							MaxConcurrentUploads: 10,
							LiveLinkName:         "live",
							GoLiveOnFinish:       true,
							StaleCleanupTimeout:  time.Minute * 30,
							Hooks:                HooksConfig{},
						},
					},
				},
				Authentication: AuthenticationProviderConfig{
					BasicAuth: &AuthenticationProviderBasicAuth{
						HTPasswdFile: "/etc/webploy/.htpasswd",
					},
				},
				Authorization: AuthorizationProviderConfig{
					PolicyFile: "/etc/webploy/policy.csv",
				},
			},
		},
		{
			name: "happy__some_simple",
			configYAML: `---
listen:
  bind_addr: ":69420"
  enable_tls: true
  tls_key: "test"
  tls_cert: "test2"

authentication:
  basic_auth: {}

sites:
  root: "/asd/asd"
  sites:
    - name: test1
      go_live_on_finish: false
      max_open: 12
      max_history: 1
      hooks:
        validator: "test1"
        pre_deploy: "test2"
        post_deploy: "test3"

    - name: test2
      max_open: 1
      max_history: 1
      max_concurrent_uploads: 300
      link_name: "asd"
      hooks:
        validator: "test4"
        pre_deploy: "test5"
        post_deploy: "test6"
`,
			expectedConfig: WebployConfig{
				Listen: ListenConfig{
					BindAddr:  ":69420",
					EnableTLS: true,
					TLSKey:    "test",
					TLSCert:   "test2",
				},
				Sites: SitesConfig{
					Root: "/asd/asd",
					Sites: []SiteConfig{
						{
							Name:                 "test1",
							MaxHistory:           1,
							MaxOpen:              12,
							MaxConcurrentUploads: 10,
							LiveLinkName:         "live",
							GoLiveOnFinish:       false,
							StaleCleanupTimeout:  time.Minute * 30,
							Hooks: HooksConfig{
								Validator:  "test1",
								PreDeploy:  "test2",
								PostDeploy: "test3",
							},
						},
						{
							Name:                 "test2",
							MaxHistory:           1,
							MaxOpen:              1,
							MaxConcurrentUploads: 300,
							LiveLinkName:         "asd",
							GoLiveOnFinish:       true,
							StaleCleanupTimeout:  time.Minute * 30,
							Hooks: HooksConfig{
								Validator:  "test4",
								PreDeploy:  "test5",
								PostDeploy: "test6",
							},
						},
					},
				},
				Authentication: AuthenticationProviderConfig{
					BasicAuth: &AuthenticationProviderBasicAuth{
						HTPasswdFile: "/etc/webploy/.htpasswd",
					},
				},
				Authorization: AuthorizationProviderConfig{
					PolicyFile: "/etc/webploy/policy.csv",
				},
			},
		},
		{
			name:        "error__invalid_yaml",
			configYAML:  `gfad ji a`,
			expectedErr: fmt.Errorf("cannot unmarshal"),
		},
		{
			name:           "error__missing_file",
			dontCreateFile: true,
			expectedErr:    fmt.Errorf("no such file or directory"), // comparing errors is hard
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lgr := zaptest.NewLogger(t)
			tmpDir := t.TempDir()
			tmpCfgFile := path.Join(tmpDir, "webploy.conf")
			t.Setenv("WEBPLOY_CONFIG", tmpCfgFile)

			if !tc.dontCreateFile {
				assert.NoErrorf(t,
					os.WriteFile(tmpCfgFile, []byte(tc.configYAML), 0o640),
					"writing temp config",
				)
			}

			cfg, err := LoadConfig(lgr)

			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedConfig, cfg)
			}

		})
	}
}
