package hooks

import (
	"errors"
	"fmt"
	"webploy-server/deployment"
	"webploy-server/deployment/info"
	"webploy-server/site"
)

type HookVars struct {
	User              string
	SiteName          string
	SitePath          string
	SiteCurrentLive   string
	DeploymentID      string
	DeploymentCreator string
	DeploymentMeta    string
	DeploymentPath    string
}

// ReadFromSite fills the SiteName, SitePath and SiteCurrentLive vars directly from site.Site
func (v *HookVars) ReadFromSite(s site.Site) (err error) {
	v.SiteName = s.GetName()
	v.SitePath = s.GetPath()
	v.SiteCurrentLive, err = s.GetLiveDeploymentID()
	return
}

// ReadFromDeployment fills the DeploymentPath, DeploymentCreator and DeploymentMeta fields from deployment.Deployment
func (v *HookVars) ReadFromDeployment(d deployment.Deployment) error {
	i, err := d.GetFullInfo()
	if err != nil {
		return err
	}
	v.DeploymentPath = d.GetPath()
	v.ReadFromDeploymentInfo(i)
	return nil
}

// ReadFromDeploymentInfo fills the DeploymentCreator and DeploymentMeta fields from info.DeploymentInfo
func (v *HookVars) ReadFromDeploymentInfo(i info.DeploymentInfo) {
	v.DeploymentCreator = i.Creator
	v.DeploymentMeta = i.Meta
}

func (v *HookVars) ReadFromSiteAndDeployment(s site.Site, d deployment.Deployment) error {
	return errors.Join(v.ReadFromSite(s), v.ReadFromDeployment(d))
}

func (v *HookVars) compileEnvvars(hook HookID) []string {
	envvars := []string{
		fmt.Sprintf("WEBPLOY_HOOK=%s", hook),
		fmt.Sprintf("WEBPLOY_USER=%s", v.User),
		fmt.Sprintf("WEBPLOY_SITE=%s", v.SiteName),
		fmt.Sprintf("WEBPLOY_SITE_PATH=%s", v.SitePath),
		fmt.Sprintf("WEBPLOY_SITE_CURRENT_LIVE=%s", v.SiteCurrentLive),
		fmt.Sprintf("WEBPLOY_DEPLOYMENT_CREATOR=%s", v.DeploymentCreator),
		fmt.Sprintf("WEBPLOY_DEPLOYMENT_META=%s", v.DeploymentMeta),
		fmt.Sprintf("WEBPLOY_DEPLOYMENT_PATH=%s", v.DeploymentPath),
		fmt.Sprintf("WEBPLOY_DEPLOYMENT_ID=%s", v.DeploymentID),
	}

	return envvars
}

func (v *HookVars) Copy() HookVars {
	return HookVars{
		User:              v.User,
		SiteName:          v.SiteName,
		SitePath:          v.SitePath,
		SiteCurrentLive:   v.SiteCurrentLive,
		DeploymentID:      v.DeploymentID,
		DeploymentCreator: v.DeploymentCreator,
		DeploymentMeta:    v.DeploymentMeta,
		DeploymentPath:    v.DeploymentPath,
	}
}
