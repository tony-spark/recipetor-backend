package db

import "embed"

//go:embed "migrations"
var EmbeddedDBFiles embed.FS
