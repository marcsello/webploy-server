package utils

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsSubDir(t *testing.T) {
	testCases := []struct {
		name        string
		argParent   string
		argSub      string
		expectedOk  bool
		expectedErr error
	}{
		{
			name:       "happy__subdir_1",
			argParent:  "/tmp/test",
			argSub:     "/tmp/test/test",
			expectedOk: true,
		},
		{
			name:       "happy__subdir_2",
			argParent:  "/tmp/test/",
			argSub:     "/tmp/test/test/",
			expectedOk: true,
		},
		{
			name:       "happy__subdir_3",
			argParent:  "/tmp/test/",
			argSub:     "/tmp/test/test",
			expectedOk: true,
		},
		{
			name:       "happy__subdir_4",
			argParent:  "/tmp/test",
			argSub:     "/tmp/test/test/",
			expectedOk: true,
		},
		{
			name:       "happy__subdir_5",
			argParent:  "/",
			argSub:     "/tmp/test/test",
			expectedOk: true,
		},
		{
			name:       "happy__subdir_6",
			argParent:  "/tmp/test",
			argSub:     "/tmp/test/../test/test",
			expectedOk: true,
		},
		{
			name:       "happy__not_subdir_1",
			argParent:  "/tmp/test",
			argSub:     "/srv/test",
			expectedOk: false,
		},
		{
			name:       "happy__not_subdir_2",
			argParent:  "/tmp/test",
			argSub:     "/tmp/../test",
			expectedOk: false,
		},
		{
			name:       "happy__not_subdir_3",
			argParent:  "/tmp/test",
			argSub:     "/tmp/../../",
			expectedOk: false,
		},
		{
			name:        "error__empty_strings",
			argParent:   "..",
			argSub:      "",
			expectedErr: fmt.Errorf("Rel: can't make"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ok, err := IsSubDir(tc.argParent, tc.argSub)
			assert.Equal(t, tc.expectedOk, ok)

			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
