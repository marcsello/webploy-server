package authentication

import (
	"fmt"
	"golang.org/x/exp/utf8string"
	"strings"
)

func ValidateUsername(name string) error {

	if name == "" {
		return fmt.Errorf("empty string")
	}

	if strings.HasPrefix(name, SystemPrefix) { // htaccess should not contain names prefixed like this
		return fmt.Errorf("%s has invalid prefix", name)
	}

	if !utf8string.NewString(name).IsASCII() { // TODO: This allows non-printable?
		return fmt.Errorf("%s contains non-ascii characters", name)
	}

	return nil

}
