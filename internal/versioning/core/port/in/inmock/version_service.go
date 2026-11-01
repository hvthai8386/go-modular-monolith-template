// Package inmock holds function-field mocks of the driving ports for adapter tests.
package inmock

import (
	"context"

	"github.com/hvthai8386/go-modular-monolith-template/internal/versioning/core/domain"
)

type VersionService struct {
	GetVersionFn func(ctx context.Context) (domain.Version, error)
}

func (m *VersionService) GetVersion(ctx context.Context) (domain.Version, error) {
	if m.GetVersionFn == nil {
		panic("inmock.VersionService: GetVersionFn is not set")
	}
	return m.GetVersionFn(ctx)
}
