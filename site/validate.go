package site

import (
	"fmt"
	"golang.org/x/exp/utf8string"
	"strings"
)

func ValidateSiteName(s string) error {

	// TODO: revisit this?

	if s == "" {
		return fmt.Errorf("empty string")
	}

	if strings.HasPrefix(s, ".") {
		return fmt.Errorf("%s has invalid prefix", s)
	}

	invalidErr := fmt.Errorf("invalid characters")

	if strings.ContainsAny(s, "/\\") {
		return invalidErr
	}

	if !utf8string.NewString(s).IsASCII() { // TODO: This allows non-printable!
		return invalidErr
	}

	return nil
}
