package site

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateSiteName(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expectedErr error
	}{
		{
			name:        "happy__simple",
			input:       "test.site.name",
			expectedErr: nil,
		},
		{
			name:        "error__empty_string",
			input:       "",
			expectedErr: fmt.Errorf("empty string"),
		},
		{
			name:        "error__prefix",
			input:       ".hidden.site",
			expectedErr: fmt.Errorf("has invalid prefix"),
		},
		{
			name:        "error__slash",
			input:       "test.site/something",
			expectedErr: fmt.Errorf("invalid characters"),
		},
		{
			name:        "error__non-ascii",
			input:       "❤️",
			expectedErr: fmt.Errorf("invalid characters"),
		},
		{
			name:        "error__non-printable_1",
			input:       "\x00",
			expectedErr: fmt.Errorf("invalid characters"),
		},
		{
			name:        "error__non-printable_2",
			input:       "\n",
			expectedErr: fmt.Errorf("invalid characters"),
		},
		{
			name:        "error__non-printable_3",
			input:       "\x7F", // del
			expectedErr: fmt.Errorf("invalid characters"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			err := ValidateSiteName(tc.input)
			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

		})
	}

}
