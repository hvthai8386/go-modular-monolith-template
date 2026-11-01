// Package migrations embeds the goose SQL migrations so a single binary can migrate and serve.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
