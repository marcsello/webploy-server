package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidatePrintableAscii(t *testing.T) {
	testCases := []struct {
		name     string
		s        string
		expected bool
	}{
		{
			name:     "happy__simple",
			s:        "test.site.name",
			expected: true,
		},
		{
			name:     "error__empty_string",
			s:        "",
			expected: false,
		},
		{
			name:     "error__non-ascii_1",
			s:        "ááááááááááááá",
			expected: false,
		},
		{
			name:     "error__non-ascii_2",
			s:        "❤️",
			expected: false,
		},
		{
			name:     "error__non-printable_1",
			s:        "\x00",
			expected: false,
		},
		{
			name:     "error__non-printable_2",
			s:        "\n",
			expected: false,
		},
		{
			name:     "error__non-printable_3",
			s:        "\x7F", // del
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, ValidatePrintableAscii(tc.s))
		})
	}
}
