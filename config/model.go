package config

import (
	"github.com/creasty/defaults"
	"time"
)

// WebployConfig is the root of the config
type WebployConfig struct {
	Listen         ListenConfig                 `yaml:"listen"`
	Authentication AuthenticationProviderConfig `yaml:"authentication"`
	Authorization  AuthorizationProviderConfig  `yaml:"authorization"`
	Sites          SitesConfig                  `yaml:"sites"`
}

type AuthenticationProviderConfig struct {
	// Currently we only plan to support BasicAuth
	BasicAuth *AuthenticationProviderBasicAuth `yaml:"basic_auth"`
}

type AuthenticationProviderBasicAuth struct {
	HTPasswdFile string `yaml:"htpasswd_file" default:"/etc/webploy/.htpasswd"`
}

func (apba *AuthenticationProviderBasicAuth) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// This function here is needed so that the defaults are set for each new element, when unmarshaling the YAML
	// the plain unmarshaler does not yet know of these fields when pre-setting the defaults, so we have to do it for every new field when they are created
	// https://stackoverflow.com/a/56080478

	err := defaults.Set(apba)
	if err != nil {
		return err
	}

	type plain AuthenticationProviderBasicAuth
	if err = unmarshal((*plain)(apba)); err != nil {
		return err
	}

	return nil
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

	MaxHistory   int    `yaml:"max_history" default:"2"` // keep this many finished deployments
	MaxOpen      int    `yaml:"max_open" default:"2"`    // how many unfinished uploads to keep open (block new ones until purged)
	LiveLinkName string `yaml:"link_name" default:"live"`

	GoLiveOnFinish bool `yaml:"go_live_on_finish" default:"true"` // automatically set a finished deployment live

	StaleCleanupTimeout time.Duration `yaml:"stale_cleanup_timeout" default:"30m"` // clean up unfinished deployments after this time

	Hooks HooksConfig `yaml:"hooks"`
}

func (sc *SiteConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// This function here is needed so that the defaults are set for each new element, when unmarshaling the YAML
	// the plain unmarshaler does not yet know of these fields when pre-setting the defaults, so we have to do it for every new field when they are created
	// https://stackoverflow.com/a/56080478

	err := defaults.Set(sc)
	if err != nil {
		return err
	}

	type plain SiteConfig
	if err = unmarshal((*plain)(sc)); err != nil {
		return err
	}

	return nil
}

type HooksConfig struct {
	// scripts
	Validator  string `yaml:"validator"`
	PreDeploy  string `yaml:"pre_deploy"`
	PostDeploy string `yaml:"post_deploy"`
}
