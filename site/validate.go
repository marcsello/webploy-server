package site

import (
	"fmt"
	"github.com/marcsello/webploy-server/utils"
	"strings"
)

func ValidateSiteName(s string) error {

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

	if !utils.ValidatePrintableAscii(s) {
		return invalidErr
	}

	return nil
}
