package site

import (
	"fmt"
	"github.com/google/uuid"
	"strings"
	"time"
)

func NewDeploymentID() string {
	u := uuid.New()
	ts := time.Now()
	return fmt.Sprintf(DeploymentIDFormatStr, ts.Format(DeploymentIDTimeFormat), u.String())
}

func ParseDeploymentID(id string) (uuid.UUID, time.Time, error) {

	parts := strings.Split(id, DeploymentIDSeparator)
	if len(parts) != 3 {
		return uuid.UUID{}, time.Time{}, fmt.Errorf("invalid part count for deployment id")
	}

	if parts[0] != DeploymentIDPrefix {
		return uuid.UUID{}, time.Time{}, fmt.Errorf("invalid prefix for deployment id")
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
	return err != nil
}
