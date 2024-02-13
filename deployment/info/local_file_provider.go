package info

import (
	"encoding/json"
	"errors"
	"github.com/marcsello/webploy-server/utils"
	"github.com/natefinch/atomic"
	"io"
	"os"
	"path"
	"syscall"
)

const InfoFileName = "info.json"

var globalInfoFileLock = utils.NewKMutex() // TODO: this is a bad solution, solve locking somehow else maybe?

type InfoProviderLocalFile struct {
	infoFilePath string
}

func NewLocalFileInfoProvider(deploymentFullPath string) *InfoProviderLocalFile {
	return &InfoProviderLocalFile{
		infoFilePath: path.Join(deploymentFullPath, InfoFileName),
	}
}

func (splf *InfoProviderLocalFile) loadData() (info DeploymentInfo, err error) {
	var file *os.File
	file, err = os.OpenFile(splf.infoFilePath, os.O_RDONLY, 0o640) // #nosec G304 G302
	if err != nil {
		return
	}
	defer func() {
		err = errors.Join(err, file.Close())
	}()

	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
	if err != nil {
		return
	}

	err = json.NewDecoder(file).Decode(&info)
	return
}

func (splf *InfoProviderLocalFile) storeData(info DeploymentInfo) error {
	// Behold, the greatest optimization known to mankind...
	reader, writer := io.Pipe()
	errChan := make(chan error)

	go func() {
		err := json.NewEncoder(writer).Encode(&info)
		if err != nil {
			errChan <- err
			return
		}
		errChan <- writer.Close()
	}()
	go func() {
		errChan <- atomic.WriteFile(splf.infoFilePath, reader)
	}()

	return errors.Join(<-errChan, <-errChan)
}

func (splf *InfoProviderLocalFile) Tx(readOnly bool, txFunc InfoTransaction) (err error) {
	// Lock the info file in this process
	globalInfoFileLock.Lock(splf.infoFilePath)
	defer globalInfoFileLock.Unlock(splf.infoFilePath)

	var info DeploymentInfo
	info, err = splf.loadData()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// file not existing, this is okay, if we want to create it for the first time
			if readOnly {
				return // only fail if read-only
			}
		} else {
			return
		}
	}

	preTxInfo := info.Copy() // should work even if there was an error loading it

	err = txFunc(&info)
	if err != nil {
		return
	}

	if !readOnly && !preTxInfo.Equals(info) {
		err = splf.storeData(info)
		if err != nil {
			return
		}
	}

	return
}
