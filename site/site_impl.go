package site

import (
	"github.com/marcsello/webploy-server/config"
	"github.com/marcsello/webploy-server/deployment"
	"github.com/marcsello/webploy-server/utils"
	"go.uber.org/zap"
	"os"
	"path"
	"sync"
)

type SiteImpl struct {
	fullPath           string // this is a read-only constant... sort of... it never changes
	deploymentsMutex   sync.RWMutex
	cfg                config.SiteConfig
	deploymentProvider deployment.Provider
	logger             *zap.Logger
}

func (s *SiteImpl) getPathForId(id string) string {
	return path.Join(s.fullPath, id)
}

func (s *SiteImpl) Init() (bool, error) {
	s.deploymentsMutex.Lock()
	defer s.deploymentsMutex.Unlock()

	// validate the name
	err := ValidateSiteName(s.cfg.Name)
	if err != nil {
		s.logger.Error("Could not init: The site has an invalid name", zap.String("Name", s.cfg.Name), zap.Error(err))
		return false, ErrSiteNameInvalid
	}

	// first, check if directory exists, and is indeed a directory
	exists, err := utils.ExistsAndDirectory(s.fullPath)
	if err != nil {
		return false, err
	}

	// if not exists, create it, and return true to indicate that this site was just created for the first time
	firstTime := false
	if !exists {
		err = os.Mkdir(s.fullPath, 0o750)
		if err != nil {
			return false, err
		}
		firstTime = true
	}

	return firstTime, nil
}

func (s *SiteImpl) GetName() string {
	return s.cfg.Name
}

func (s *SiteImpl) GetPath() string {
	return s.fullPath
}

func (s *SiteImpl) GetConfig() config.SiteConfig {
	return s.cfg
}

func (s *SiteImpl) listDeploymentIDs() ([]string, error) {
	var ids []string

	entries, err := os.ReadDir(s.fullPath)
	if err != nil {
		// may fail when the directory is missing
		return nil, err
	}

	for _, e := range entries {
		if e.IsDir() && IsDeploymentIDValid(e.Name()) {
			ids = append(ids, e.Name())
		}
	}

	return ids, nil
}

func (s *SiteImpl) ListDeploymentIDs() ([]string, error) {
	s.deploymentsMutex.RLock()
	defer s.deploymentsMutex.RUnlock()
	return s.listDeploymentIDs()
}

func (s *SiteImpl) GetDeployment(id string) (deployment.Deployment, error) {
	if !IsDeploymentIDValid(id) {
		return nil, ErrInvalidID
	}
	fullPath := s.getPathForId(id)

	s.deploymentsMutex.RLock()
	defer s.deploymentsMutex.RUnlock()

	exists, err := utils.ExistsAndDirectory(fullPath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrDeploymentNotExists
	}

	s.logger.Debug("Loading existing deployment", zap.String("deploymentID", id), zap.String("deploymentFullPath", fullPath))
	return s.deploymentProvider.LoadDeployment(fullPath)
}

// IterDeployments is a safe iterator for iterating trough deployments,
// it holds a read-lock for the deployments while iterating, so the list of deployments does not change
func (s *SiteImpl) IterDeployments(iter DeploymentIterator) error {
	s.deploymentsMutex.RLock()
	defer s.deploymentsMutex.RUnlock()

	list, err := s.listDeploymentIDs()
	if err != nil {
		s.logger.Error("Failed to list deployments", zap.Error(err))
		return err
	}

	var liveID string
	liveID, err = s.readLiveDeploymentIDFromSymlink()
	if err != nil {
		s.logger.Error("Failed to read deployment ID from Symlink", zap.Error(err))
		return err
	}

	for _, id := range list {
		deploymentPath := path.Join(s.fullPath, id)
		var d deployment.Deployment
		d, err = s.deploymentProvider.LoadDeployment(deploymentPath)
		if err != nil {
			return err
		}
		var cont bool
		cont, err = iter(id, d, liveID == id)
		if !cont {
			break
		}
	}

	return nil
}

func (s *SiteImpl) CreateNewDeployment(creator, meta string) (string, deployment.Deployment, error) {
	s.deploymentsMutex.Lock()
	defer s.deploymentsMutex.Unlock()

	var newID string
	var err error
	var newDeploymentFullPath string

	for i := 0; i < 10; i++ {
		newID = NewDeploymentID()
		newDeploymentFullPath = s.getPathForId(newID)

		err = os.Mkdir(newDeploymentFullPath, 0o750)
		if err != nil {
			if os.IsExist(err) {
				s.logger.Debug("Generated colliding entry. Retrying...", zap.Int("retryCounter", i))
				continue // retry
			}
			return "", nil, err
		}
		// success
		break
	}
	s.logger.Info("Initializing new deployment", zap.String("deploymentID", newID), zap.String("deploymentFullPath", newDeploymentFullPath))
	var d deployment.Deployment
	d, err = s.deploymentProvider.InitDeployment(newDeploymentFullPath, creator, meta)
	return newID, d, err
}

func (s *SiteImpl) deleteDeployment(id string) error {
	fullPath := s.getPathForId(id)
	underDeletePath := fullPath + ".delete"

	// don't allow deleting the active deployment
	sID, err := s.readLiveDeploymentIDFromSymlink()
	if err == nil && id == sID {
		return ErrDeploymentLive
	}

	// if it wasn't the active, and the id is valid, we can try to delete it.
	// it could still fail to if the deployment does not exist

	// first, we "atomically" move it out of the way, so subsequent get requests will fail with deployment not existing
	// this also breaks the symlink
	s.logger.Debug("Renaming deployment before deletion", zap.String("deploymentID", id), zap.String("deploymentFullPath", fullPath), zap.String("underDeletePath", underDeletePath))
	err = os.Rename(fullPath, underDeletePath)
	if err != nil {
		return err
	}

	// we then do removing in the background
	go func() {
		s.logger.Debug("deployment delete started in the background", zap.String("deploymentID", id), zap.String("underDeletePath", underDeletePath))
		e := os.RemoveAll(underDeletePath)
		if e != nil {
			s.logger.Error("failed to delete deployment folder", zap.String("deploymentID", id), zap.Error(e), zap.String("underDeletePath", underDeletePath))
		}
	}()

	return nil
}

func (s *SiteImpl) DeleteDeployment(id string) error {
	// check if id is valid
	if !IsDeploymentIDValid(id) {
		return ErrInvalidID
	}

	// lock
	s.deploymentsMutex.Lock()
	defer s.deploymentsMutex.Unlock()

	return s.deleteDeployment(id)
}

func (s *SiteImpl) SetLiveDeploymentID(id string) error {
	if !IsDeploymentIDValid(id) {
		return ErrInvalidID
	}
	symlinkFullPath := path.Join(s.fullPath, s.cfg.LiveLinkName)
	tmpSymlinkFullPath := symlinkFullPath + ".new"

	// lock
	s.deploymentsMutex.Lock()
	defer s.deploymentsMutex.Unlock()

	// existence check
	dp, err := s.deploymentProvider.LoadDeployment(s.getPathForId(id))
	if err != nil {
		return err
	}

	// check if the deployment can go live
	var finished bool
	finished, err = dp.IsFinished()
	if err != nil {
		return err
	}
	if !finished {
		return ErrDeploymentNotFinished
	}

	// all good, proceed with going live
	s.logger.Debug("All checks completed, proceeding to create symlinks", zap.String("deploymentID", id), zap.String("tmpSymlinkFullPath", tmpSymlinkFullPath), zap.String("symlinkFullPath", symlinkFullPath))

	// clean any leftover link
	err = os.Remove(tmpSymlinkFullPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	// create "new" link
	err = os.Symlink(id, tmpSymlinkFullPath) // make it a relative link
	if err != nil {
		return err
	}

	// "atomic" replace
	err = os.Rename(tmpSymlinkFullPath, symlinkFullPath)
	if err != nil {
		return err
	}

	return nil // success
}

func (s *SiteImpl) readLiveDeploymentIDFromSymlink() (string, error) {
	symlinkFullPath := path.Join(s.fullPath, s.cfg.LiveLinkName)
	dest, err := os.Readlink(symlinkFullPath)
	if err != nil {
		return "", err
	}

	return path.Base(dest), nil
}

func (s *SiteImpl) GetLiveDeploymentID() (string, error) {
	s.deploymentsMutex.RLock()
	defer s.deploymentsMutex.RUnlock()

	id, err := s.readLiveDeploymentIDFromSymlink()
	if err != nil {
		return "", err
	}

	if !IsDeploymentIDValid(id) {
		return "", ErrInvalidID
	}

	return id, nil
}
