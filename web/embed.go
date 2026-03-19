package web

import "embed"

// Static holds all files under web/static/, embedded into the binary at
// compile time. The server mounts this at /static/ via http.FS so no
// filesystem dependency exists at runtime.
//
//go:embed static
var Static embed.FS
