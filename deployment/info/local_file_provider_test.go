package info

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
	"time"
)

func TestNewLocalFileInfoProvider_StoreNew(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name string
		info DeploymentInfo
	}{
		{
			name: "simple_1",
			info: DeploymentInfo{
				Creator:        "test",
				CreatedAt:      now,
				State:          DeploymentStateOpen,
				FinishedAt:     nil,
				LastActivityAt: now,
			},
		},
		{
			name: "simple_2",
			info: DeploymentInfo{
				Creator:        "test",
				CreatedAt:      now,
				State:          DeploymentStateOpen,
				FinishedAt:     &now,
				LastActivityAt: now,
			},
		},
		{
			name: "simple_3",
			info: DeploymentInfo{
				Creator:        "test",
				CreatedAt:      now,
				State:          DeploymentStateFinished,
				FinishedAt:     nil,
				LastActivityAt: now,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testInfoFilePath := path.Join(tmpDir, InfoFileName)
			assert.NoFileExists(t, testInfoFilePath)

			lfip := NewLocalFileInfoProvider(tmpDir)
			assert.NoFileExists(t, testInfoFilePath)

			err := lfip.storeData(tc.info)
			assert.NoError(t, err)

			assert.FileExists(t, testInfoFilePath)

			// Check equality from one side
			marshaledJson, err := json.Marshal(tc.info)
			assert.NoError(t, err)

			writtenJson, err := os.ReadFile(testInfoFilePath)
			assert.NoError(t, err)

			assert.JSONEq(t, string(marshaledJson), string(writtenJson))

			// Check equality from the other side
			var writtenInfo DeploymentInfo
			err = json.Unmarshal(writtenJson, &writtenInfo)
			assert.NoError(t, err)

			assert.True(t, writtenInfo.Equals(tc.info))
			assert.True(t, tc.info.Equals(writtenInfo))
		})
	}

}

func TestNewLocalFileInfoProvider_StoreNewError(t *testing.T) {

	tmpDir := t.TempDir()
	testInfoFilePath := path.Join(tmpDir, InfoFileName)

	_ = os.Mkdir(testInfoFilePath, 0o750)
	assert.DirExists(t, testInfoFilePath)

	lfip := NewLocalFileInfoProvider(tmpDir)

	err := lfip.storeData(DeploymentInfo{})
	assert.Error(t, err)
	// it is not easy to check for specific errors in golang
	assert.Contains(t, err.Error(), "cannot replace")

}

func TestNewLocalFileInfoProvider_LoadError(t *testing.T) {

	tmpDir := t.TempDir()
	testInfoFilePath := path.Join(tmpDir, InfoFileName)

	assert.NoFileExists(t, testInfoFilePath)

	lfip := NewLocalFileInfoProvider(tmpDir)

	_, err := lfip.loadData()
	assert.Error(t, err)
	assert.ErrorIs(t, err, os.ErrNotExist)

	// ---

	_ = os.Mkdir(testInfoFilePath, 0o750)
	assert.DirExists(t, testInfoFilePath)

	_, err = lfip.loadData()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is a directory")

}

func TestNewLocalFileInfoProvider_Load(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name string
		info DeploymentInfo
	}{
		{
			name: "simple_1",
			info: DeploymentInfo{
				Creator:        "test",
				CreatedAt:      now,
				State:          DeploymentStateOpen,
				FinishedAt:     nil,
				LastActivityAt: now,
			},
		},
		{
			name: "simple_2",
			info: DeploymentInfo{
				Creator:        "test",
				CreatedAt:      now,
				State:          DeploymentStateOpen,
				FinishedAt:     &now,
				LastActivityAt: now,
			},
		},
		{
			name: "simple_3",
			info: DeploymentInfo{
				Creator:        "test",
				CreatedAt:      now,
				State:          DeploymentStateFinished,
				FinishedAt:     nil,
				LastActivityAt: now,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testInfoFilePath := path.Join(tmpDir, InfoFileName)

			file, _ := os.Create(testInfoFilePath)
			err := json.NewEncoder(file).Encode(tc.info)
			assert.NoError(t, err)
			_ = file.Close()

			assert.FileExists(t, testInfoFilePath)

			lfip := NewLocalFileInfoProvider(tmpDir)

			loadedInfo, err := lfip.loadData()
			assert.NoError(t, err)

			assert.True(t, loadedInfo.Equals(tc.info))
			assert.True(t, tc.info.Equals(loadedInfo))
		})
	}

}
