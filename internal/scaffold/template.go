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
}

const fileMode = fs.FileMode(0o644)

func always(ProjectProfile) bool { return true }

func platformIs(id PlatformID) func(ProjectProfile) bool {
	return func(p ProjectProfile) bool { return p.Platform == id }
}

func dockerOn(p ProjectProfile) bool { return p.Docker }
