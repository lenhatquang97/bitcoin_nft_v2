package db

import (
	"embed"
	_ "embed"
)

//go:embed migration/*.up.sql
var sqlSchemas embed.FS
