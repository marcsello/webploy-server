package site

import (
	"errors"
	"fmt"
	"os"
	"path"
	"sync"
	"webploy-server/config"
	"webploy-server/deployment"
	"webploy-server/utils"
)

type SiteImpl struct {
	fullPath           string
	deploymentsMutex   sync.RWMutex
	cfg                config.SiteConfig
	deploymentProvider deployment.Provider
}


func (s *SiteImpl) Init() (bool, error) {
	s.deploymentsMutex.Lock()
	defer s.deploymentsMutex.Unlock()

	// first, check if directory exists, and is indeed a directory
	exists, err := utils.ExistsAndDirectory(s.fullPath)
	if err != nil {
		return false, err
	}

	// if not exists, create it, and return true to indicate that this site was just created for the first time
	firstTime := false
	if !exists {
		err = os.Mkdir(s.fullPath, 0o770)
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

func (s *SiteImpl) GetConfig() config.SiteConfig {
	return s.cfg
}

func (s *SiteImpl) ListDeploymentIDs() ([]string, error) {
	s.deploymentsMutex.RLock()
	defer s.deploymentsMutex.RUnlock()

	ids := []string

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

func (s *SiteImpl) GetDeployment(id string) (deployment.Deployment, error) {
	if !IsDeploymentIDValid(id) {
		return nil, ErrInvalidID
	}

	s.deploymentsMutex.RLock()
	defer s.deploymentsMutex.RUnlock()

	return s.deploymentProvider.LoadDeployment(id)
}

func (s *SiteImpl) CreateNewDeployment(creator string) (deployment.Deployment, error) {
	s.deploymentsMutex.Lock()
	defer s.deploymentsMutex.Unlock()

	var newID string
	var err error
	var dp deployment.Deployment

	// TODO: enforce maximum open deployments

	for i := 0; i < 10; i++ {
		newID = NewDeploymentID()
		dp, err = s.deploymentProvider.CreateDeployment(newID, creator) // it should create it's own folder and stuff

		if !errors.Is(err, deployment.ErrDeploymentAlreadyExists) {
			// retry only on id collision
			break
		}
	}

	return dp, err
}

func (s *SiteImpl) DeleteDeployment(id string) error {
	// check if id is valid
	if !IsDeploymentIDValid(id) {
		return ErrInvalidID
	}

	// lock
	s.deploymentsMutex.Lock()
	defer s.deploymentsMutex.Unlock()

	// don't allow deleting the active deployment
	sID, err := s.readLiveDeploymentIDFromSymlink()
	if err == nil && id == sID {
		return ErrDeploymentLive
	}

	// if it wasn't the active, and the id is valid, we can try to delete it.
	// it could still fail to if the deployment does not exist
	return s.deploymentProvider.DeleteDeployment(id)
}

func (s *SiteImpl) SetLiveDeploymentID(id string) error {
	if !IsDeploymentIDValid(id) {
		return ErrInvalidID
	}
	symlinkFullPath := path.Join(s.fullPath, s.cfg.LinkName)
	tmpSymlinkFullPath := symlinkFullPath + ".new"

	// lock
	s.deploymentsMutex.Lock()
	defer s.deploymentsMutex.Unlock()

	// existence check
	dp, err := s.deploymentProvider.LoadDeployment(id)
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
		return fmt.Errorf("deployment not finished")
	}

	// all good, proceed with going live

	// clean any leftover link
	err = os.Remove(tmpSymlinkFullPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	// create "new" link
	err = os.Symlink(tmpSymlinkFullPath, id) // make it a relative link
	if err != nil {
		return err
	}

	// "atomic" replace
	err = os.Rename(symlinkFullPath, tmpSymlinkFullPath)
	if err != nil {
		return err
	}

	return nil // success
}

func (s *SiteImpl) readLiveDeploymentIDFromSymlink() (string, error) {
	symlinkFullPath := path.Join(s.fullPath, s.cfg.LinkName)
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