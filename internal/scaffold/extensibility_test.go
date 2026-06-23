package scaffold_test

import (
	"testing"

	"github.com/sgaunet/scaffold/internal/scaffold"
)

// TestNewTemplateAppearsAutomatically: registering a template makes it appear
// in Applicable() (used by both generate and list) with no other change
// (FR-015, SC-008). This is the extensibility guarantee.
func TestNewTemplateAppearsAutomatically(t *testing.T) {
	t.Parallel()
	reg := scaffold.NewRegistry()
	p := scaffold.ProjectProfile{ProjectName: "demo", Binary: "demo", MainPath: "./cmd/demo"}

	before := len(reg.Applicable(p))

	reg.Register(scaffold.Template{
		Name:     "editorconfig",
		Source:   "templates/editorconfig.tmpl", // source need not exist to be listed
		DestTmpl: ".editorconfig",
		Applies:  func(scaffold.ProjectProfile) bool { return true },
	})

	after := reg.Applicable(p)
	if len(after) != before+1 {
		t.Fatalf("Applicable count = %d, want %d", len(after), before+1)
	}
	found := false
	for _, tmpl := range after {
		if tmpl.Name == "editorconfig" {
			found = true
		}
	}
	if !found {
		t.Fatal("newly registered template did not appear in Applicable()")
	}
}
