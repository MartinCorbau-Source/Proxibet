package migrations

import "embed"

// Files contains every SQL migration embedded in the API binary.
//
//go:embed *.sql
var Files embed.FS
