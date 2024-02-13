package hooks

import (
	"fmt"
	"github.com/marcsello/webploy-server/deployment"
	"github.com/marcsello/webploy-server/deployment/info"
	"github.com/marcsello/webploy-server/site"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHookVars_Copy(t *testing.T) {
	v1 := HookVars{
		User:              "test1",
		SiteName:          "test2",
		SitePath:          "test3",
		SiteCurrentLive:   "test4",
		DeploymentID:      "test5",
		DeploymentCreator: "test6",
		DeploymentMeta:    "test7",
		DeploymentPath:    "test8",
	}
	v2 := v1.Copy()
	assert.Equal(t, v1, v2)
}

func TestHookVars_ReadFromDeployment(t *testing.T) {
	// happy
	d := new(deployment.MockDeployment)
	d.On("GetFullInfo").Return(info.DeploymentInfo{
		Creator: "test1",
		Meta:    "test2",
	}, nil)
	d.On("GetPath").Return("test3")

	v := HookVars{}
	e := v.ReadFromDeployment(d)
	assert.NoError(t, e)

	assert.Equal(t, v.DeploymentCreator, "test1")
	assert.Equal(t, v.DeploymentMeta, "test2")
	assert.Equal(t, v.DeploymentPath, "test3")

	// error
	testErr := fmt.Errorf("test error")
	d2 := new(deployment.MockDeployment)
	d2.On("GetFullInfo").Return(info.DeploymentInfo{}, testErr)
	d2.On("GetPath").Return("test1")

	v2 := HookVars{}
	e = v2.ReadFromDeployment(d2)
	assert.Error(t, e)
	assert.Equal(t, testErr, e)
}

func TestHookVars_ReadFromSite(t *testing.T) {
	// happy
	s := new(site.MockSite)
	s.On("GetName").Return("test1")
	s.On("GetPath").Return("test2")
	s.On("GetLiveDeploymentID").Return("test3", nil)

	v := HookVars{}
	e := v.ReadFromSite(s)
	assert.NoError(t, e)

	assert.Equal(t, v.SiteName, "test1")
	assert.Equal(t, v.SitePath, "test2")
	assert.Equal(t, v.SiteCurrentLive, "test3")

	// error
	testErr := fmt.Errorf("test")
	s2 := new(site.MockSite)
	s2.On("GetName").Return("test1")
	s2.On("GetPath").Return("test2")
	s2.On("GetLiveDeploymentID").Return("", testErr)

	v2 := HookVars{}
	e = v2.ReadFromSite(s2)
	assert.Error(t, e)
	assert.Equal(t, testErr, e)
}

func TestHookVars_ReadFromDeploymentInfo(t *testing.T) {
	i := info.DeploymentInfo{
		Creator: "test1",
		Meta:    "test2",
	}
	v := HookVars{}
	v.ReadFromDeploymentInfo(i)
	assert.Equal(t, i.Creator, v.DeploymentCreator)
	assert.Equal(t, i.Meta, v.DeploymentMeta)
}

func TestHookVars_ReadFromSiteAndDeployment(t *testing.T) {
	testErr := fmt.Errorf("test error")
	dGood := new(deployment.MockDeployment)
	dGood.On("GetFullInfo").Return(info.DeploymentInfo{
		Creator: "test1",
		Meta:    "test2",
	}, nil)
	dGood.On("GetPath").Return("test3")
	dBad := new(deployment.MockDeployment)
	dBad.On("GetFullInfo").Return(info.DeploymentInfo{}, testErr)
	dBad.On("GetPath").Return("")

	sGood := new(site.MockSite)
	sGood.On("GetName").Return("test1")
	sGood.On("GetPath").Return("test2")
	sGood.On("GetLiveDeploymentID").Return("test3", nil)
	sBad := new(site.MockSite)
	sBad.On("GetName").Return("")
	sBad.On("GetPath").Return("")
	sBad.On("GetLiveDeploymentID").Return("", testErr)
	// happy
	testCases := []struct {
		name          string
		argDeployment deployment.Deployment
		argSite       site.Site
		expectErr     bool
	}{
		{
			name:          "happy__both_good",
			argDeployment: dGood,
			argSite:       sGood,
			expectErr:     false,
		},
		{
			name:          "error__site_bad",
			argDeployment: dGood,
			argSite:       sBad,
			expectErr:     true,
		},
		{
			name:          "error__deployment_bad",
			argDeployment: dBad,
			argSite:       sGood,
			expectErr:     true,
		},
		{
			name:          "error__both_bad",
			argDeployment: dBad,
			argSite:       sBad,
			expectErr:     true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := HookVars{}
			err := v.ReadFromSiteAndDeployment(tc.argSite, tc.argDeployment)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHookVars_compileEnvvars(t *testing.T) {
	v := HookVars{
		User:              "test1",
		SiteName:          "test2",
		SitePath:          "test3",
		SiteCurrentLive:   "test4",
		DeploymentID:      "test5",
		DeploymentCreator: "test6",
		DeploymentMeta:    "test7",
		DeploymentPath:    "test8",
	}

	expectedEnvvars := []string{
		"WEBPLOY_USER=test1",
		"WEBPLOY_SITE=test2",
		"WEBPLOY_SITE_PATH=test3",
		"WEBPLOY_SITE_CURRENT_LIVE=test4",
		"WEBPLOY_DEPLOYMENT_ID=test5",
		"WEBPLOY_DEPLOYMENT_CREATOR=test6",
		"WEBPLOY_DEPLOYMENT_META=test7",
		"WEBPLOY_DEPLOYMENT_PATH=test8",
		"WEBPLOY_HOOK=test9",
	}

	compiledEnvvars := v.compileEnvvars("test9")

	assert.ElementsMatch(t, expectedEnvvars, compiledEnvvars)
}
