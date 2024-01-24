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
	fullPath      string
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
func (d *DeploymentImpl) Init(creator string) error {
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
		return nil
	})
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

func (d *DeploymentImpl) Finish() error {
	// TODO: don't allow finishing while uploads are still pending
	// Store the pending uploads in the info?
	return d.infoProvider.Tx(false, func(i *info.DeploymentInfo) error {

		if i.IsFinished() {
			return ErrDeploymentFinished
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

	// update info
	err = d.infoProvider.Tx(false, func(i *info.DeploymentInfo) error {

		if i.IsFinished() {
			return ErrDeploymentFinished
		}

		i.LastActivityAt = time.Now()

		return nil

	})
	if err != nil {
		return err
	}
	// TODO: record pending upload, enforce max concurrent uploads

	d.logger.Debug("Creating destination file", zap.String("destPath", destPath))
	var file *os.File
	file, err = os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return err
	}

	d.logger.Debug("Receiving file", zap.String("destPath", destPath))
	var bytesWritten int64
	bytesWritten, err = ctxio.Copy(ctx, file, stream)
	if err != nil {
		// if anything goes wrong, close and delete the file...
		err2 := file.Close()
		err3 := os.Remove(destPath)
		return errors.Join(err, err2, err3)
	}
	d.logger.Info("Successfully written file", zap.Int64("bytesWritten", bytesWritten), zap.String("relpath", relpath), zap.String("destPath", destPath))

	return file.Close()
}
