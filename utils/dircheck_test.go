package utils

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func TestExistsAndDirectory(t *testing.T) {
	testCases := []struct {
		name        string
		preTestFn   func(string)
		argPath     string // relative to tmp dir
		expectedOk  bool
		expectedErr error
	}{
		{
			name: "happy__simple",
			preTestFn: func(tmpDir string) {
				_ = os.Mkdir(path.Join(tmpDir, "test"), 0o777)
			},
			argPath:     "test",
			expectedOk:  true,
			expectedErr: nil,
		},
		{
			name:        "happy__missing",
			preTestFn:   func(string) {},
			argPath:     "test",
			expectedOk:  false,
			expectedErr: nil,
		},
		{
			name: "error_not_dir",
			preTestFn: func(tmpDir string) {
				f, _ := os.Create(path.Join(tmpDir, "test"))
				_ = f.Close()
			},
			argPath:     "test",
			expectedOk:  false,
			expectedErr: ErrNotDir,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tc.preTestFn(tmpDir)

			ok, err := ExistsAndDirectory(path.Join(tmpDir, tc.argPath))

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
