package utils

import (
	"errors"
	"os"
)

var ErrNotDir = errors.New("is not a directory")

func ExistsAndDirectory(path string) (bool, error) {
	var exists = true
	file, err := os.Open(path) // #nosec G304
	if err != nil {
		if os.IsNotExist(err) {
			exists = false
		} else {
			return false, err
		}
	}

	if exists {
		var fileInfo os.FileInfo
		fileInfo, err = file.Stat()
		if err != nil {
			return false, err
		}

		if !fileInfo.IsDir() {
			return false, ErrNotDir
		}
	}

	return exists, nil
}
