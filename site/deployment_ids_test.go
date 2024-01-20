package site

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGeneratorValid(t *testing.T) {
	testdeploymentid := NewDeploymentID()
	result := IsDeploymentIDValid(testdeploymentid)
	assert.True(t, result)
}
