package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"open-end/internal/council"
	"path/filepath"
	"strings"
)

func runTree(args []string, out io.Writer) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "-h" {
		_, err := fmt.Fprintln(out, "council tree init -input <completed batch/trial> -out <new tree> [-variant control] [-window 50000]\ncouncil tree show -tree <tree> [-json]\ncouncil tree select -tree <tree> -branches b000002,b000003\ncouncil tree grow -tree <tree> [-branches b000002,b000003] [-proposals] [-ticks 20000 -every 1000 -window 10000 -workers 16]\ncouncil tree export -tree <tree> [-out <report.html>]\n\nGrow continues selected branches with unchanged rules. With -proposals, every parent must have its own round/response.json; paired controls are always retained. Export produces an offline interactive tree and command builder.")
		if err == nil {
			_, err = fmt.Fprintln(out, "\ncouncil tree archive -tree <tree> [-apply] [-json] [-min-ticks 10000 -min-survival 0.75 -min-persistence 0.75 -bins 1 -neighbors 3]\nArchive records immutable novelty/Pareto decisions. -apply saves recommendations without starting simulation.")
		}
		return err
	}
	command := args[0]
	fs := flag.NewFlagSet("tree "+command, flag.ContinueOnError)
	fs.SetOutput(out)
	var dir, input, dest, variant, branches string
	var asJSON, proposals, apply bool
	archiveOpts := council.DefaultArchiveOptions()
	opts := council.TrialOptions{Ticks: 20000, Every: 1000, Window: 10000, Workers: 16}
	switch command {
	case "init":
		fs.StringVar(&input, "input", "", "completed batch or trial")
		fs.StringVar(&dest, "out", "", "new tree directory")
		fs.StringVar(&variant, "variant", "", "trial variant to import")
		fs.Uint64Var(&opts.Window, "window", 50000, "evidence window")
	case "show", "select", "grow", "export", "archive":
		fs.StringVar(&dir, "tree", "", "tree directory")
		switch command {
		case "archive":
			fs.BoolVar(&apply, "apply", false, "save recommended branches as the continuation selection")
			fs.BoolVar(&asJSON, "json", false, "machine-readable archive decision")
			fs.Uint64Var(&archiveOpts.MinTicks, "min-ticks", 10000, "minimum stable observation window")
			fs.Float64Var(&archiveOpts.MinSurvival, "min-survival", .75, "minimum fraction of surviving worlds")
			fs.Float64Var(&archiveOpts.MinPersistence, "min-persistence", .75, "minimum fraction of productive time blocks")
			fs.IntVar(&archiveOpts.Bins, "bins", 1, "behavior cell bins per log-rate octave")
			fs.IntVar(&archiveOpts.Neighbors, "neighbors", 3, "nearest archive neighbors for novelty")
		case "show":
			fs.BoolVar(&asJSON, "json", false, "machine-readable tree")
		case "select":
			fs.StringVar(&branches, "branches", "", "comma-separated branch IDs to retain for continuation")
		case "grow":
			fs.StringVar(&branches, "branches", "", "override saved selection for this run")
			fs.BoolVar(&proposals, "proposals", false, "read each parent's round/response.json and run paired proposals")
			fs.IntVar(&opts.Ticks, "ticks", 20000, "additional ticks per world")
			fs.IntVar(&opts.Every, "every", 1000, "report interval")
			fs.Uint64Var(&opts.Window, "window", 10000, "analysis window, at most ticks")
			fs.IntVar(&opts.Workers, "workers", 16, "parallel worlds and GOMAXPROCS")
		case "export":
			fs.StringVar(&dest, "out", "", "HTML destination; default tree/index.html")
		}
	default:
		return fmt.Errorf("unknown tree command %q", command)
	}
	if err := fs.Parse(args[1:]); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected arguments")
	}
	if command == "init" {
		if input == "" || dest == "" {
			return fmt.Errorf("provide -input and -out")
		}
		if err := council.InitTree(input, dest, variant, opts.Window); err != nil {
			return err
		}
		_, err := fmt.Fprintln(out, "Tree created:", dest)
		return err
	}
	if dir == "" {
		return fmt.Errorf("provide -tree")
	}
	ids := []string{}
	if branches != "" {
		for _, id := range strings.Split(branches, ",") {
			ids = append(ids, strings.TrimSpace(id))
		}
	}
	switch command {
	case "archive":
		a, err := council.ArchiveTree(dir, archiveOpts, apply)
		if err != nil {
			return err
		}
		if asJSON {
			enc := json.NewEncoder(out)
			enc.SetIndent("", "  ")
			return enc.Encode(a)
		}
		_, err = fmt.Fprint(out, council.ArchiveText(a))
		return err
	case "select":
		if err := council.SelectTree(dir, ids); err != nil {
			return err
		}
		_, err := fmt.Fprintln(out, "Selected:", strings.Join(ids, ","))
		return err
	case "grow":
		children, err := council.GrowTree(dir, ids, proposals, opts, out)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(out, "Selected children:", strings.Join(children, ","))
		return err
	case "export":
		if dest == "" {
			dest = filepath.Join(dir, "index.html")
		}
		if err := council.ExportTree(dir, dest); err != nil {
			return err
		}
		_, err := fmt.Fprintln(out, "Open:", dest)
		return err
	default:
		v, err := council.ReadTree(dir)
		if err != nil {
			return err
		}
		if asJSON {
			enc := json.NewEncoder(out)
			enc.SetIndent("", "  ")
			return enc.Encode(v)
		}
		_, err = fmt.Fprint(out, council.TreeText(v))
		return err
	}
}
