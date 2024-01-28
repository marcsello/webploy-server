package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSitesConfig_GetConfigForSite(t *testing.T) {

	sc := SitesConfig{
		Sites: []SiteConfig{
			{
				Name:     "test1",
				LinkName: "test1",
			},
			{
				Name:     "test2",
				LinkName: "test2",
			},
		},
	}

	s, ok := sc.GetConfigForSite("test1")
	assert.True(t, ok)
	assert.Equal(t, "test1", s.LinkName)

	s, ok = sc.GetConfigForSite("test2")
	assert.True(t, ok)
	assert.Equal(t, "test2", s.LinkName)

	_, ok = sc.GetConfigForSite("test3")
	assert.False(t, ok)

}
