package info

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path"
	"syscall"
	"webploy-server/utils"
)

const InfoFileName = "info.json"

var globalInfoFileLock = utils.NewKmutex()

type InfoProviderLocalFile struct {
	infoFilePath string
}

func NewLocalFileInfoProvider(deploymentFullPath string) *InfoProviderLocalFile {
	return &InfoProviderLocalFile{
		infoFilePath: path.Join(deploymentFullPath, InfoFileName),
	}
}

func (splf *InfoProviderLocalFile) Tx(readOnly bool, txFunc InfoTransaction) (err error) {
	// Lock the info file in this process
	globalInfoFileLock.Lock(splf.infoFilePath)
	defer globalInfoFileLock.Unlock(splf.infoFilePath)

	flags := os.O_CREATE
	if readOnly {
		flags |= os.O_RDONLY
	} else {
		flags |= os.O_RDWR
	}
	var file *os.File
	file, err = os.OpenFile(splf.infoFilePath, flags, 0o640)
	defer func() {
		err = errors.Join(err, file.Close())
	}()

	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
	if err != nil {
		return
	}

	var infoBytes []byte
	infoBytes, err = io.ReadAll(file)

	var info DeploymentInfo

	if len(infoBytes) > 0 { // Note: this will silently fail if the file became empty for some reason, TODO: fix it
		err = json.Unmarshal(infoBytes, &info)
		if err != nil {
			return
		}
	} else {
		info = DeploymentInfo{} // use empty struct then
	}

	preTxInfo := info.Copy()

	err = txFunc(&info)
	if err != nil {
		return
	}

	if !readOnly && !preTxInfo.Equals(info) {
		// overwriting the file
		_, err = file.Seek(0, 0)
		if err != nil {
			return
		}

		err = file.Truncate(0)
		if err != nil {
			return
		}

		err = json.NewEncoder(file).Encode(info)
		if err != nil {
			return
		}
	}

	return
}
