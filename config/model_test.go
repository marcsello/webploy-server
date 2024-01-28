package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSitesConfig_GetConfigForSite(t *testing.T) {

	sc := SitesConfig{
		Sites: []SiteConfig{
			{
				Name:         "test1",
				LiveLinkName: "test1",
			},
			{
				Name:         "test2",
				LiveLinkName: "test2",
			},
		},
	}

	s, ok := sc.GetConfigForSite("test1")
	assert.True(t, ok)
	assert.Equal(t, "test1", s.LiveLinkName)

	s, ok = sc.GetConfigForSite("test2")
	assert.True(t, ok)
	assert.Equal(t, "test2", s.LiveLinkName)

	_, ok = sc.GetConfigForSite("test3")
	assert.False(t, ok)

}
