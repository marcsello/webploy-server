package site

import (
	"fmt"
	"github.com/google/uuid"
	"strings"
	"time"
)

const DeploymentIDSeparator = "_"
const DeploymentIDPrefix = "deployment"
const DeploymentIDFormatStr = DeploymentIDPrefix + DeploymentIDSeparator + "%s" + DeploymentIDSeparator + "%s"
const DeploymentIDTimeFormat = "2006-01-02-15-04-05"

const DeploymentPathDeleteSuffix = ".delete"

func NewDeploymentID() string {
	u := uuid.New()
	ts := time.Now()
	return fmt.Sprintf(DeploymentIDFormatStr, ts.Format(DeploymentIDTimeFormat), u.String())
}

func ParseDeploymentID(id string) (uuid.UUID, time.Time, error) {

	if strings.HasSuffix(id, DeploymentPathDeleteSuffix) {
		return uuid.UUID{}, time.Time{}, fmt.Errorf("this ID have a path suffix: %s", DeploymentPathDeleteSuffix)
	}

	parts := strings.Split(id, DeploymentIDSeparator)
	if len(parts) != 3 {
		return uuid.UUID{}, time.Time{}, fmt.Errorf("invalid part count for deployment id: %d", len(parts))
	}

	if parts[0] != DeploymentIDPrefix {
		return uuid.UUID{}, time.Time{}, fmt.Errorf("invalid prefix for deployment id: %s", parts[0])
	}

	ts, err := time.Parse(DeploymentIDTimeFormat, parts[1])
	if err != nil {
		return uuid.UUID{}, time.Time{}, err
	}

	u, err := uuid.Parse(parts[2])
	if err != nil {
		return uuid.UUID{}, time.Time{}, err
	}

	return u, ts, nil

}

func IsDeploymentIDValid(id string) bool {
	_, _, err := ParseDeploymentID(id)
	return err == nil
}
