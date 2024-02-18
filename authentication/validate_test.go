package authentication

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateUsername(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expectedErr error
	}{
		{
			name:        "happy__simple",
			input:       "test username",
			expectedErr: nil,
		},
		{
			name:        "error__empty_string",
			input:       "",
			expectedErr: fmt.Errorf("empty string"),
		},
		{
			name:        "error__prefix",
			input:       "_system",
			expectedErr: fmt.Errorf("has invalid prefix"),
		},
		{
			name:        "error__prefix",
			input:       "❤️",
			expectedErr: fmt.Errorf("contains non-ascii characters"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			err := ValidateUsername(tc.input)
			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

		})
	}

}
