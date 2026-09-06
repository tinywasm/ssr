//go:build !wasm

package sitec_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"webtyp.com/sitec"
)

func TestExtract_RealFinder_ToleratesDependencyModules(t *testing.T) {
	if testing.Short() {
		t.Skip("spawns `go run`; skipped with -short")
	}
	root := t.TempDir()

	goMod := `module example.com/app

go 1.25.2

require (
	webtyp.com/widget v0.3.0
	webtyp.com/css v0.3.0
	webtyp.com/fmt v0.25.5
	webtyp.com/js v0.0.4
)
` + webtypReplaces(t)
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a subpackage with a widget implementing RenderCSS()
	if err := os.MkdirAll(filepath.Join(root, "config"), 0755); err != nil {
		t.Fatal(err)
	}

	cssContent := `//go:build !wasm

package config

import (
	"webtyp.com/css"
	"webtyp.com/widget"
)

type MyWidget struct{}

func (m *MyWidget) WidgetName() widget.Name { return widget.Name("my-widget") }
func (m *MyWidget) WidgetKind() widget.Kind { return widget.Region }
func (m *MyWidget) RenderCSS() *css.Stylesheet {
	return css.NewStylesheet(css.Raw(".my-widget{padding:0}"))
}
`
	if err := os.WriteFile(filepath.Join(root, "config", "css.go"), []byte(cssContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Run go mod tidy to resolve dependencies
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed: %v, output: %s", err, string(out))
	}

	e := sitec.New(root) // finder REAL, sin Seed
	assets, err := e.ExtractModule(root)
	if err != nil {
		t.Fatalf("extraction failed with a real finder: %v", err)
	}
	if assets == nil {
		t.Fatal("expected non-nil assets")
	}
	if !strings.Contains(assets.CSS, "my-widget") {
		t.Fatalf("expected CSS from the widget's RenderCSS(), got: %q", assets.CSS)
	}
}
