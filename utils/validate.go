package utils

import (
	"regexp"
)

var printableRe = regexp.MustCompile(`^[[:print:]]+$`)

func ValidatePrintableAscii(s string) bool {
	return printableRe.MatchString(s)
}
