package council

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

//go:embed tree.html
var treePage string

// ExportTree emits a self-contained offline report. UI selections only build
// explicit CLI commands; opening a report never mutates or launches a world.
func ExportTree(dir, dest string) error {
	if strings.ToLower(filepath.Ext(dest)) != ".html" {
		return fmt.Errorf("report destination must use .html")
	}
	v, err := ReadTree(dir)
	if err != nil {
		return err
	}
	archive, err := LatestArchive(dir)
	if err != nil {
		return err
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	t, err := template.New("tree").Parse(treePage)
	if err != nil {
		return err
	}
	var b bytes.Buffer
	model := struct {
		Tree         TreeView
		Directory    string
		Archive      *ArchiveDecision
		ArchiveStale bool
	}{v, abs, archive, archive != nil && archive.TreeHash != jsonHash(v.Nodes)}
	if err = t.Execute(&b, model); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(dest), ".report-*")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if _, err = f.Write(b.Bytes()); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(name, dest)
}
