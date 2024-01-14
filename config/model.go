package config

import "time"

// WebployConfig is the root of the config
type WebployConfig struct {
	Listen         ListenConfig
	Sites          SitesConfig
	Authentication AuthenticationProviderConfig
	Authorization  AuthorizationProviderConfig
}

type AuthenticationProviderConfig struct {
	// Currently we only plan to support BasicAuth
	BasicAuth *AuthenticationProviderBasicAuth
}

type AuthenticationProviderBasicAuth struct {
	HTPasswdFile string
}

type AuthorizationProviderConfig struct {
	PolicyFile string
}

type ListenConfig struct {
	BindAddr  string
	EnableTLS bool
	TLSKey    string
	TLSCert   string
}

type SitesConfig struct {
	Root  string
	Sites []SiteConfig
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
	Name string // this will be the "resource name" in the authorization

	MaxHistory int
	MaxOpen    int
	LinkName   string

	DeployOnFinish bool

	StaleCleanupTimeout time.Duration // clean up unfinished deployments after this time

	// scripts
	Validator  string
	PreDeploy  string
	PostDeploy string
}
