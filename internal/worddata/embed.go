package worddata

import "embed"

// Words contains curated English and Spanish word lists baked into the binary.
//
//go:embed en es
var Words embed.FS
