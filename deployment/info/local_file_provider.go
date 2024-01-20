package info

import (
	"path"
	"reflect"
)

const InfoFileName = "info.json"

type InfoProviderLocalFile struct {
	infoFilePath string
}

func NewLocalFileInfoProvider(deploymentFullPath string) (*InfoProviderLocalFile, error) {
	return &InfoProviderLocalFile{
		infoFilePath: path.Join(deploymentFullPath, InfoFileName),
	}, nil
}

func (splf *InfoProviderLocalFile) Tx(readonly bool, txFunc InfoTransaction) error {
	var err error

	var info *DeploymentInfo

	info, err = splf.readState()
	if err != nil {
		return err
	}
	if info == nil {
		// should be an error
	}

	err = txFunc(info)
	if err != nil {
		return err
	}

	if currentState == nil || !reflect.DeepEqual(newState, currentState) {
		splf.writeState(newState)
	}

	return nil
}
