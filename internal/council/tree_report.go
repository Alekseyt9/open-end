package council

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"open-end/internal/observer"
	"os"
	"path/filepath"
	"strings"
)

//go:embed tree.html
var treePage string

type environmentView struct {
	*observer.EnvironmentWindow
	Copies map[string]uint64 `json:"copies_by_actor"`
}

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
	variation := map[string]map[string]*observer.VariationWindow{}
	environment := map[string]map[string]*environmentView{}
	for _, n := range v.Nodes {
		for _, w := range n.Worlds {
			if w.CopyModel == "" && w.Environment == "" {
				continue
			}
			var e Evidence
			if err := readJSON(filepath.Join(nodeRound(dir, n.ID), w.ID+".evidence.json"), &e); err != nil {
				return err
			}
			if variation[n.ID] == nil {
				variation[n.ID] = map[string]*observer.VariationWindow{}
			}
			variation[n.ID][w.ID] = e.Summary.Variation
			if environment[n.ID] == nil {
				environment[n.ID] = map[string]*environmentView{}
			}
			if e.Summary.Environment != nil {
				v := &environmentView{EnvironmentWindow: e.Summary.Environment, Copies: map[string]uint64{}}
				for _, a := range e.Summary.Activity {
					v.Copies[a.Hash] = a.Copies
				}
				environment[n.ID][w.ID] = v
			}
		}
	}
	model := struct {
		Tree         TreeView
		Directory    string
		Archive      *ArchiveDecision
		ArchiveStale bool
		Variation    map[string]map[string]*observer.VariationWindow
		Environment  map[string]map[string]*environmentView
	}{v, abs, archive, archive != nil && archive.TreeHash != jsonHash(v.Nodes), variation, environment}
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
