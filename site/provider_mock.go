package site

import (
	"github.com/stretchr/testify/mock"
)

// MockProvider is a mock implementation of the Provider interface.
type MockProvider struct {
	mock.Mock
}

// GetSite mocks the GetSite method of the Provider interface.
func (m *MockProvider) GetSite(name string) (Site, bool) {
	args := m.Called(name)
	return args.Get(0).(Site), args.Bool(1)
}

// GetAllSiteNames mocks the GetAllSiteNames method of the Provider interface.
func (m *MockProvider) GetAllSiteNames() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

// GetNewSiteNamesSinceInit mocks the GetNewSiteNamesSinceInit method of the Provider interface.
func (m *MockProvider) GetNewSiteNamesSinceInit() []string {
	args := m.Called()
	return args.Get(0).([]string)
}
