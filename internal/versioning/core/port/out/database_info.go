// Package out holds the driven ports: what the versioning module needs from the outside world.
package out

import "context"

type DatabaseInfoReader interface {
	ServerVersion(ctx context.Context) (string, error)
	SchemaVersion(ctx context.Context) (int64, error)
}
