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
	s.deploymentsMutex.RLock()
	defer s.deploymentsMutex.RUnlock()

	if !IsDeploymentIDValid(id) {
		return nil, fmt.Errorf("invalid deployment id")
	}

	expectedPath := path.Join(s.fullPath, id)
	return s.deploymentProvider.LoadExistingDeployment(expectedPath, id, s.cfg)
}

func (s *SiteImpl) CreateNewDeployment(creator string) (deployment.Deployment, error) {
	s.deploymentsMutex.Lock()
	defer s.deploymentsMutex.Unlock()

	var newID, expectedPath string
	var err error
	var dp deployment.Deployment

	// TODO: enforce maximum open deployments

	for i := 0; i < 10; i++ {
		newID = NewDeploymentID()
		expectedPath = path.Join(s.fullPath, newID)
		dp, err = s.deploymentProvider.CreateNewDeployment(expectedPath, newID, s.cfg, creator) // it should create it's own folder and stuff

		if !errors.Is(err, deployment.ErrDeploymentAlreadyExists) {
			// retry only on id collision
			break
		}
	}

	return dp, err
}

func (s *SiteImpl) SetLiveDeploymentID(id string) error {
	s.deploymentsMutex.Lock()
	defer s.deploymentsMutex.Unlock()
	if !IsDeploymentIDValid(id) {
		return fmt.Errorf("invalid id")
	}

	symlinkFullPath := path.Join(s.fullPath, s.cfg.LinkName)
	tmpSymlinkFullPath := symlinkFullPath + ".new"

	err := os.Remove(tmpSymlinkFullPath) // clean any leftover link
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

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

func (s *SiteImpl) GetLiveDeploymentID() (string, error) {
	s.deploymentsMutex.RLock()
	defer s.deploymentsMutex.RUnlock()

	symlinkFullPath := path.Join(s.fullPath, s.cfg.LinkName)

	dest, err := os.Readlink(symlinkFullPath)
	if err != nil {
		return "", err
	}

	basename := path.Base(dest)
	if !IsDeploymentIDValid(basename) {
		return "", fmt.Errorf("link resolves to invalid id")
	}

	return basename, nil
}