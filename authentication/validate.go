package authentication

import (
	"fmt"
	"github.com/marcsello/webploy-server/utils"
	"strings"
)

func ValidateUsername(name string) error {

	if name == "" {
		return fmt.Errorf("empty string")
	}

	if strings.HasPrefix(name, SystemPrefix) { // htaccess should not contain names prefixed like this
		return fmt.Errorf("%s has invalid prefix", name)
	}

	if !utils.ValidatePrintableAscii(name) {
		return fmt.Errorf("%s contains non-ascii characters", name)
	}

	return nil

}
