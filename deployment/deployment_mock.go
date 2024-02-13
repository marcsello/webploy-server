package deployment

import (
	"context"
	"github.com/marcsello/webploy-server/deployment/info"
	"github.com/stretchr/testify/mock"
	"io"
	"time"
)

// MockDeployment is a mock implementation of the Deployment interface.
type MockDeployment struct {
	mock.Mock
}

// GetPath mocks the GetPath method of the Deployment interface.
func (m *MockDeployment) GetPath() string {
	args := m.Called()
	return args.String(0)
}

// AddFile mocks the AddFile method of the Deployment interface.
func (m *MockDeployment) AddFile(ctx context.Context, relpath string, stream io.ReadCloser) error {
	args := m.Called(ctx, relpath, stream)
	return args.Error(0)
}

// IsFinished mocks the IsFinished method of the Deployment interface.
func (m *MockDeployment) IsFinished() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

// Finish mocks the Finish method of the Deployment interface.
func (m *MockDeployment) Finish() error {
	args := m.Called()
	return args.Error(0)
}

// Creator mocks the Creator method of the Deployment interface.
func (m *MockDeployment) Creator() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

// LastActivity mocks the LastActivity method of the Deployment interface.
func (m *MockDeployment) LastActivity() (time.Time, error) {
	args := m.Called()
	return args.Get(0).(time.Time), args.Error(1)
}

// GetFullInfo mocks the GetFullInfo method of the Deployment interface.
func (m *MockDeployment) GetFullInfo() (info.DeploymentInfo, error) {
	args := m.Called()
	return args.Get(0).(info.DeploymentInfo), args.Error(1)
}
