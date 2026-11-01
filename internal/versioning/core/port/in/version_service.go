// Package in holds the driving ports: what the outside world may ask the versioning module to do.
package in

import (
	"context"

	"github.com/hvthai8386/go-modular-monolith-template/internal/versioning/core/domain"
)

type VersionService interface {
	GetVersion(ctx context.Context) (domain.Version, error)
}
