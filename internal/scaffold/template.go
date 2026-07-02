package scaffold

import "io/fs"

// Template describes one generatable file: its embedded source, its (rendered)
// destination, and an applicability predicate (always / platform== / docker).
type Template struct {
	Name     string
	Source   string // path inside the embedded FS, e.g. "templates/mise.toml.tmpl"
	DestTmpl string // destination path (itself rendered) relative to the target dir
	Applies  func(ProjectProfile) bool
	Mode     fs.FileMode

	// DisableIfExists, if non-empty, is a path relative to the target dir whose
	// presence causes Dest to be suffixed with DisableSuffix, neutralizing the
	// generated file for its runner without skipping generation entirely.
	// Directory-vs-file is not distinguished (a stray file at this path is
	// treated the same as a directory).
	DisableIfExists string
	DisableSuffix   string // e.g. ".disabled"; only meaningful if DisableIfExists != ""
}

const fileMode = fs.FileMode(0o644)

func always(ProjectProfile) bool { return true }

func platformIs(id PlatformID) func(ProjectProfile) bool {
	return func(p ProjectProfile) bool { return p.Platform == id }
}

func dockerOn(p ProjectProfile) bool { return p.Docker }
