package domain

import "testing"

func TestNewAppVersion(t *testing.T) {
	valid := []string{"1.0.0", "v1.0.0", "0.0.0-dev", "2.1.3-rc.1+build.7"}
	for _, raw := range valid {
		if _, err := NewAppVersion(raw); err != nil {
			t.Errorf("NewAppVersion(%q) returned error: %v", raw, err)
		}
	}

	invalid := []string{"", "1", "1.0", "latest", "1.0.0 "}
	for _, raw := range invalid {
		if _, err := NewAppVersion(raw); err == nil {
			t.Errorf("NewAppVersion(%q) accepted an invalid version", raw)
		}
	}
}
