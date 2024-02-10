package deployment

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"io"
	"jayconrod.com/ctxio"
	"os"
	"path"
	"time"
	"webploy-server/config"
	"webploy-server/deployment/info"
	"webploy-server/utils"
)

const ContentSubDirName = "_content"

type DeploymentImpl struct {
	infoProvider  info.InfoProvider // <- store all state info here, as the Deployment objects generally live for a single request and they are not shared
	fullPath      string            // full path of the deployment (used as unique id for the deployment)
	contentSubDir string
	siteConfig    config.SiteConfig
	logger        *zap.Logger
}

func NewDeployment(fullPath string, siteConfig config.SiteConfig, logger *zap.Logger) *DeploymentImpl {
	return &DeploymentImpl{
		infoProvider:  info.NewLocalFileInfoProvider(fullPath),
		fullPath:      fullPath,
		contentSubDir: path.Join(fullPath, ContentSubDirName),
		siteConfig:    siteConfig,
		logger:        logger,
	}
}

// Init lays down the basic structure of the deployment. This should be only called when initializing a new deployment
func (d *DeploymentImpl) Init(creator, meta string) error {
	err := os.Mkdir(d.contentSubDir, 0o750)
	if err != nil {
		return err
	}
	return d.infoProvider.Tx(true, func(i *info.DeploymentInfo) error {
		now := time.Now()
		i.State = info.DeploymentStateOpen
		i.CreatedAt = now
		i.LastActivityAt = now
		i.Creator = creator
		i.Meta = meta
		return nil
	})
}

func (d *DeploymentImpl) GetPath() string {
	return d.fullPath
}

func (d *DeploymentImpl) IsFinished() (bool, error) {
	var isFinished bool
	err := d.infoProvider.Tx(true, func(i *info.DeploymentInfo) error {
		isFinished = i.IsFinished()
		return nil
	})
	return isFinished, err
}

func (d *DeploymentImpl) Creator() (string, error) {
	var creator string
	err := d.infoProvider.Tx(true, func(i *info.DeploymentInfo) error {
		creator = i.Creator
		return nil
	})
	return creator, err
}

func (d *DeploymentImpl) LastActivity() (time.Time, error) {
	var lastActivity time.Time
	err := d.infoProvider.Tx(true, func(i *info.DeploymentInfo) error {
		lastActivity = i.LastActivityAt
		return nil
	})
	return lastActivity, err
}

func (d *DeploymentImpl) GetFullInfo() (info.DeploymentInfo, error) {
	var i_ info.DeploymentInfo
	err := d.infoProvider.Tx(true, func(i *info.DeploymentInfo) error {
		i_ = i.Copy()
		return nil
	})
	return i_, err
}

// because deployment objects are short-lived objects, we have to put this here...
// also, this is purely runtime info, would not make sense to store it in the state
var pendingUploads = utils.NewKCounter() // TODO: maybe set this up with the provider?

func (d *DeploymentImpl) Finish() error {
	return d.infoProvider.Tx(false, func(i *info.DeploymentInfo) error {

		if i.IsFinished() {
			return ErrDeploymentFinished
		}

		if pendingUploads.Get(d.fullPath) > 0 { // we check this inside the transaction, to reduce the race condition likeliness
			return ErrUploadPending
		}

		i.State = info.DeploymentStateFinished
		now := time.Now()
		i.FinishedAt = &now
		i.LastActivityAt = now

		return nil

	})
}

func (d *DeploymentImpl) AddFile(ctx context.Context, relpath string, stream io.ReadCloser) error {
	// prepare and validate path
	if path.IsAbs(relpath) {
		return ErrUploadInvalidPath
	}
	destPath := path.Join(d.contentSubDir, relpath)
	subdir, err := utils.IsSubDir(d.contentSubDir, destPath)
	if err != nil {
		return err
	}
	if !subdir {
		return ErrUploadInvalidPath
	}
	destSubdir := path.Dir(destPath)
	if destSubdir != "" {
		subdir, err = utils.IsSubDir(d.contentSubDir, destPath)
		if err != nil {
			return err
		}
		if !subdir {
			return ErrUploadInvalidPath
		}
	}

	d.logger.Debug("Path checks complete", zap.String("destPath", destPath), zap.String("destSubdir", destSubdir))

	// Limit concurrent uploads
	// we do this before checking for finished deployment,
	// so it is not possible to set the deployment finished AFTER we checked it
	uploadCnt := pendingUploads.Incr(d.fullPath) // use deployment path as lock name
	defer pendingUploads.Dec(d.fullPath)
	if d.siteConfig.MaxConcurrentUploads != 0 && uploadCnt > d.siteConfig.MaxConcurrentUploads {
		err = ErrTooManyConcurrentUploads
		d.logger.Debug("Max concurrent upload limit has reached", zap.Error(err))
		return err
	}
	d.logger.Debug("Concurrent uploads limit is not reached", zap.Uint("uploadCnt", uploadCnt), zap.Uint("MaxConcurrentUploads", d.siteConfig.MaxConcurrentUploads))

	// update info
	err = d.infoProvider.Tx(false, func(i *info.DeploymentInfo) error {

		if i.IsFinished() {
			return ErrDeploymentFinished
		}

		i.LastActivityAt = time.Now()

		return nil

	})
	if err != nil {
		// not doing error logging because ErrDeploymentFinished is not an error
		return err
	}
	d.logger.Debug("State file updated")

	// Ensure containing dir
	err = os.MkdirAll(destSubdir, 0o640)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			d.logger.Error("Error while ensuring containing directory", zap.Error(err), zap.String("destSubdir", destSubdir))
			return err
		}
		d.logger.Debug("Containing dir already exists", zap.String("destSubdir", destSubdir))
	}

	// Receive file
	d.logger.Debug("Creating destination file", zap.String("destPath", destPath))
	var file *os.File
	file, err = os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640) // #nosec G304 G302
	if err != nil {
		d.logger.Error("Error creating destination file", zap.Error(err))
		return err
	}

	d.logger.Debug("Receiving file...", zap.String("destPath", destPath))
	var bytesWritten int64
	bytesWritten, err = ctxio.Copy(ctx, file, stream)
	if err != nil {
		// if anything goes wrong, close and delete the file...
		err2 := file.Close()
		err3 := os.Remove(destPath)
		return errors.Join(err, err2, err3)
	}
	d.logger.Info("Successfully written file",
		zap.Int64("bytesWritten", bytesWritten),
		zap.String("relpath", relpath),
		zap.String("destPath", destPath),
	)

	return file.Close()
}
