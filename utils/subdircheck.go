package utils

import (
	"os"
	"path/filepath"
	"strings"
)

func IsSubDir(parent, sub string) (bool, error) {
	// source: https://stackoverflow.com/a/62529061
	up := ".." + string(os.PathSeparator)

	// path-comparisons using filepath.Abs don't work reliably according to docs (no unique representation).
	rel, err := filepath.Rel(parent, sub)
	if err != nil {
		return false, err
	}
	if !strings.HasPrefix(rel, up) && rel != ".." {
		return true, nil
	}
	return false, nil
}
