package domain

import (
	"fmt"
	"regexp"
)

// Version is what the service reports about itself and the storage it runs on.
type Version struct {
	App      AppVersion
	Go       string
	Database string
	Schema   int64
}

// AppVersion is a semantic version such as "1.4.0" or "0.0.0-dev".
type AppVersion string

var semverPattern = regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$`)

func NewAppVersion(raw string) (AppVersion, error) {
	if !semverPattern.MatchString(raw) {
		return "", fmt.Errorf("app version %q is not a semantic version", raw)
	}
	return AppVersion(raw), nil
}

func (v AppVersion) String() string {
	return string(v)
}
