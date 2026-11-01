// Package outmock holds function-field mocks of the driven ports for use case tests.
package outmock

import "context"

type DatabaseInfoReader struct {
	ServerVersionFn func(ctx context.Context) (string, error)
	SchemaVersionFn func(ctx context.Context) (int64, error)
}

func (m *DatabaseInfoReader) ServerVersion(ctx context.Context) (string, error) {
	if m.ServerVersionFn == nil {
		panic("outmock.DatabaseInfoReader: ServerVersionFn is not set")
	}
	return m.ServerVersionFn(ctx)
}

func (m *DatabaseInfoReader) SchemaVersion(ctx context.Context) (int64, error) {
	if m.SchemaVersionFn == nil {
		panic("outmock.DatabaseInfoReader: SchemaVersionFn is not set")
	}
	return m.SchemaVersionFn(ctx)
}
