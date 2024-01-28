package site

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGeneratorValid(t *testing.T) {
	testdeploymentid := NewDeploymentID()
	result := IsDeploymentIDValid(testdeploymentid)
	assert.True(t, result)
}

func TestParseDeploymentID(t *testing.T) {
	testCases := []struct {
		name          string
		argID         string
		expectedUUID  uuid.UUID
		expectedTs    time.Time
		expectedError error
	}{
		{
			name:         "happy__valid_1",
			argID:        "deployment_2012-01-02-15-04-05_0c319b16-8e3f-4064-96d1-e219bf5fa910",
			expectedUUID: uuid.MustParse("0c319b16-8e3f-4064-96d1-e219bf5fa910"),
			expectedTs:   time.Date(2012, 1, 2, 15, 4, 5, 0, time.UTC),
		},
		{
			name:         "happy__valid_2",
			argID:        "deployment_2024-12-12-15-04-05_3505e43d-8aef-41f2-8b30-b05a590d50e0",
			expectedUUID: uuid.MustParse("3505e43d-8aef-41f2-8b30-b05a590d50e0"),
			expectedTs:   time.Date(2024, 12, 12, 15, 4, 5, 0, time.UTC),
		},
		{
			name:          "error__have_suffix_1",
			argID:         "deployment_2024-12-12-15-04-05_3505e43d-8aef-41f2-8b30-b05a590d50e0.delete",
			expectedError: fmt.Errorf("this ID have a path suffix"),
		},
		{
			name:          "error__have_suffix_2",
			argID:         "deployment_.delete",
			expectedError: fmt.Errorf("this ID have a path suffix"),
		},
		{
			name:          "error__empty_string",
			argID:         "",
			expectedError: fmt.Errorf("invalid part count for deployment id"),
		},
		{
			name:          "error__extra_part",
			argID:         "deployment_2024-12-12-15-04-05_3505e43d-8aef-41f2-8b30-b05a590d50e0_hello_world",
			expectedError: fmt.Errorf("invalid part count for deployment id"),
		},
		{
			name:          "error__missing_part",
			argID:         "deployment_2024-12-12-15-04-05",
			expectedError: fmt.Errorf("invalid part count for deployment id"),
		},
		{
			name:          "error__invalid_prefix_1",
			argID:         "depasd_2024-12-12-15-04-05_3505e43d-8aef-41f2-8b30-b05a590d50e0",
			expectedError: fmt.Errorf("invalid prefix for deployment id"),
		},
		{
			name:          "error__invalid_prefix_2",
			argID:         "_2024-12-12-15-04-05_3505e43d-8aef-41f2-8b30-b05a590d50e0",
			expectedError: fmt.Errorf("invalid prefix for deployment id"),
		},
		{
			name:          "error__invalid_date_1",
			argID:         "deployment_2012-69-02-15-96-05_0c319b16-8e3f-4064-96d1-e219bf5fa910",
			expectedError: fmt.Errorf("parsing time"),
		},
		{
			name:          "error__invalid_date_2",
			argID:         "deployment_26-05_0c319b16-8e3f-4064-96d1-e219bf5fa910",
			expectedError: fmt.Errorf("parsing time"),
		},
		{
			name:          "error__invalid_date_3",
			argID:         "deployment__0c319b16-8e3f-4064-96d1-e219bf5fa910",
			expectedError: fmt.Errorf("parsing time"),
		},
		{
			name:          "error__invalid_uuid_1",
			argID:         "deployment_2012-01-02-15-04-05_0c319b16-8e3f-4064-96d1-e219bf5fa910-123",
			expectedError: fmt.Errorf("invalid UUID length"),
		},
		{
			name:          "error__invalid_uuid_2",
			argID:         "deployment_2012-01-02-15-04-05_0c319b16-8e3f-4064-96d1-hhhh",
			expectedError: fmt.Errorf("invalid UUID length"),
		},
		{
			name:          "error__invalid_uuid_3",
			argID:         "deployment_2012-01-02-15-04-05_",
			expectedError: fmt.Errorf("invalid UUID length"),
		},
		{
			name:          "error__invalid_uuid_4",
			argID:         "deployment_2012-01-02-15-04-05_0c319b16-8e3f-4064-96d1-epljbf5fa910",
			expectedError: fmt.Errorf("invalid UUID format"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u, ts, err := ParseDeploymentID(tc.argID)
			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError.Error())
				assert.False(t, IsDeploymentIDValid(tc.argID))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedUUID, u)
				assert.Equal(t, tc.expectedTs, ts)
				assert.True(t, IsDeploymentIDValid(tc.argID))
			}
		})
	}
}
