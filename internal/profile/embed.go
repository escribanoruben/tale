package profile

import "embed"

// Files contains generation profiles baked into the binary.
//
//go:embed *.toml
var Files embed.FS
