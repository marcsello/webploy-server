package utils

import (
	"fmt"
	"os"
)

func ExistsAndDirectory(path string) (bool, error) {
	var exists = true
	file, err := os.Open(path)
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
			return false, fmt.Errorf("%s exists but not a directory", path)
		}
	}

	return exists, nil
}
