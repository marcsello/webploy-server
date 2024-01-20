package config

import "time"

// WebployConfig is the root of the config
type WebployConfig struct {
	Listen         ListenConfig                 `yaml:"listen"`
	Sites          SitesConfig                  `yaml:"sites"`
	Authentication AuthenticationProviderConfig `yaml:"authentication"`
	Authorization  AuthorizationProviderConfig  `yaml:"authorization"`
}

type AuthenticationProviderConfig struct {
	// Currently we only plan to support BasicAuth
	BasicAuth *AuthenticationProviderBasicAuth `yaml:"basic_auth"`
}

type AuthenticationProviderBasicAuth struct {
	HTPasswdFile string `yaml:"htpasswd_file" default:"/etc/webploy/htpasswd"`
}

type AuthorizationProviderConfig struct {
	PolicyFile string `yaml:"policy_file" default:"/etc/webploy/policy.csv"`
}

type ListenConfig struct {
	BindAddr  string `yaml:"bind_addr" default:":8000"`
	EnableTLS bool   `yaml:"enable_tls" default:"false"`
	TLSKey    string `yaml:"tls_key"`
	TLSCert   string `yaml:"tls_cert"`
}

type SitesConfig struct {
	Root  string       `yaml:"root" default:"/var/www"`
	Sites []SiteConfig `yaml:"sites" default:"[]"`
}

func (sc *SitesConfig) GetConfigForSite(name string) (SiteConfig, bool) {

	for _, s := range sc.Sites {
		if s.Name == name {
			return s, true
		}
	}

	return SiteConfig{}, false
}

type SiteConfig struct {
	Name string `yaml:"name"` // this will be the "resource name" in the authorization

	MaxHistory int    `yaml:"max_history" default:"2"` // keep this many finished deployments
	MaxOpen    int    `yaml:"max_open" default:"2"`    // how many unfinished uploads to keep open (block new ones until purged)
	LinkName   string `yaml:"link_name" default:"live"`

	GoLiveOnFinish bool `yaml:"go_live_on_finish" default:"true"` // automatically set a finished deployment live

	StaleCleanupTimeout time.Duration `yaml:"stale_cleanup_timeout" default:"30m"` // clean up unfinished deployments after this time

	Hooks HooksConfig `yaml:"hooks"`
}

type HooksConfig struct {
	// scripts
	Validator  string `yaml:"validator"`
	PreDeploy  string `yaml:"pre_deploy"`
	PostDeploy string `yaml:"post_deploy"`
}
