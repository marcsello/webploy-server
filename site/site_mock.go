package site

import (
	"github.com/marcsello/webploy-server/config"
	"github.com/marcsello/webploy-server/deployment"
	"github.com/stretchr/testify/mock"
)

// MockSite is a mock implementation of the Site interface.
type MockSite struct {
	mock.Mock
}

// GetName mocks the GetName method of the Site interface.
func (m *MockSite) GetName() string {
	args := m.Called()
	return args.String(0)
}

// GetPath mocks the GetPath method of the Site interface.
func (m *MockSite) GetPath() string {
	args := m.Called()
	return args.String(0)
}

// GetConfig mocks the GetConfig method of the Site interface.
func (m *MockSite) GetConfig() config.SiteConfig {
	args := m.Called()
	return args.Get(0).(config.SiteConfig)
}

// ListDeploymentIDs mocks the ListDeploymentIDs method of the Site interface.
func (m *MockSite) ListDeploymentIDs() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

// GetDeployment mocks the GetDeployment method of the Site interface.
func (m *MockSite) GetDeployment(id string) (deployment.Deployment, error) {
	args := m.Called(id)
	return args.Get(0).(deployment.Deployment), args.Error(1)
}

// IterDeployments mocks the IterDeployments method of the Site interface.
func (m *MockSite) IterDeployments(iter DeploymentIterator) error {
	args := m.Called(iter)
	return args.Error(0)
}

// CreateNewDeployment mocks the CreateNewDeployment method of the Site interface.
func (m *MockSite) CreateNewDeployment(creator, meta string) (string, deployment.Deployment, error) {
	args := m.Called(creator, meta)
	return args.String(0), args.Get(1).(deployment.Deployment), args.Error(2)
}

// DeleteDeployment mocks the DeleteDeployment method of the Site interface.
func (m *MockSite) DeleteDeployment(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// SetLiveDeploymentID mocks the SetLiveDeploymentID method of the Site interface.
func (m *MockSite) SetLiveDeploymentID(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// GetLiveDeploymentID mocks the GetLiveDeploymentID method of the Site interface.
func (m *MockSite) GetLiveDeploymentID() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}
