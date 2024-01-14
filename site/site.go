package site

import (
	"fmt"
	"os"
	"path"
	"sync"
	"webploy-server/config"
	"webploy-server/deployment"
)

type Site struct {
	fullPath         string
	deploymentsMutex sync.RWMutex
	cfg              config.SiteConfig

	deploymentProvider deployment.Provider
}

func isExistsAndDirectory(path string) (bool, error) {
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
			return false, fmt.Errorf("exists but not a directory")
		}
	}

	return exists, nil
}

func (s *Site) Init() (bool, error) {
	s.deploymentsMutex.Lock()
	defer s.deploymentsMutex.Unlock()

	// first, check if directory exists, and is indeed a directory
	exists, err := isExistsAndDirectory(s.fullPath)
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

func (s *Site) GetName() string {
	return s.cfg.Name
}

func (s *Site) GetConfig() config.SiteConfig {
	return s.cfg
}

func (s *Site) ListDeploymentIDs() ([]string, error) {
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

func (s *Site) GetDeployment(id string) (deployment.Deployment, error) {
	s.deploymentsMutex.RLock()
	defer s.deploymentsMutex.RUnlock()

	if !IsDeploymentIDValid(id) {
		return nil, fmt.Errorf("invalid deployment id")
	}

	expectedPath := path.Join(s.fullPath, id)
	exists, err := isExistsAndDirectory(expectedPath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("deployment does not exists")
	}

	return s.deploymentProvider.LoadExistingDeployment(expectedPath, id, s.cfg)
}

func (s *Site) CreateNewDeployment(creator string) (deployment.Deployment, error) {
	s.deploymentsMutex.Lock()
	defer s.deploymentsMutex.Unlock()

	var newID, expectedPath string
	var success bool
	var err error

	// TODO: enforce maximum open deployments

	for i := 0; i < 10; i++ {
		newID = NewDeploymentID()
		expectedPath = path.Join(s.fullPath, newID)
		var exists bool
		exists, err = isExistsAndDirectory(expectedPath)
		if !exists && err == nil {
			success = true
			break
		}
	}

	if !success {
		return nil, fmt.Errorf("failed to allocate new deployment ID")
	}

	return s.deploymentProvider.CreateNewDeployment(expectedPath, newID, s.cfg, creator)
}

func (s *Site) SetLiveDeploymentID(id string) error {
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

func (s *Site) GetLiveDeploymentID() (string, error) {
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
