package scaffold

import "embed"

// templatesFS is the single embed entry point; every generatable file ships
// inside the binary (FR-009) while staying a reviewable file in the repo (FR-008).
//
//go:embed all:templates
var templatesFS embed.FS
