package main

import (
	"flag"
	"fmt"
	"io"
	"open-end/internal/council"
	"os"
	"path/filepath"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "council:", err)
		os.Exit(1)
	}
}
func run(args []string, out io.Writer) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		_, err := fmt.Fprintln(out, "council prepare -input <completed experiment> -out <new round> [-window 50000]\ncouncil check -round <round> [-response <response.json>]\ncouncil trial -round <round> -out <new trial> [-response <response.json>] [-ticks 20000 -every 1000 -window 10000 -workers 16]\n\nAI работает через файлы: prepare → заполнить response.json в чате → check → trial. Исходные миры не изменяются.")
		return err
	}
	command := args[0]
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	fs.SetOutput(out)
	var input, round, response, dest, variant string
	var window uint64
	opts := council.TrialOptions{}
	switch command {
	case "prepare":
		fs.StringVar(&input, "input", "", "completed experiment directory")
		fs.StringVar(&dest, "out", "", "new round directory")
		fs.Uint64Var(&window, "window", 50000, "recent evidence window")
		fs.StringVar(&variant, "variant", "", "select a variant when preparing from a completed council trial")
	case "check", "trial":
		fs.StringVar(&round, "round", "", "prepared round directory")
		fs.StringVar(&response, "response", "", "AI response JSON, defaults to round/response.json")
		if command == "trial" {
			fs.StringVar(&dest, "out", "", "new trial directory")
			fs.IntVar(&opts.Ticks, "ticks", 20000, "additional ticks per branch")
			fs.IntVar(&opts.Every, "every", 1000, "reporting interval")
			fs.IntVar(&opts.Workers, "workers", 16, "parallel workers and GOMAXPROCS")
			fs.Uint64Var(&opts.Window, "window", 10000, "analysis window")
		}
	default:
		return fmt.Errorf("unknown command %q", command)
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
	if command == "prepare" {
		if input == "" || dest == "" || window == 0 {
			return fmt.Errorf("provide -input, -out and a positive -window")
		}
		r, err := council.PrepareVariant(input, dest, window, variant)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(out, "Prepared %d worlds: %s\nRead %s; write %s\n", len(r.Worlds), r.ID, filepath.Join(dest, "brief.md"), filepath.Join(dest, "response.json"))
		return err
	}
	if round == "" {
		return fmt.Errorf("provide -round")
	}
	if response == "" {
		response = filepath.Join(round, "response.json")
	}
	if command == "check" {
		c, err := council.Check(round, response)
		if err != nil {
			return err
		}
		return council.Review(out, c)
	}
	if dest == "" {
		return fmt.Errorf("provide -out")
	}
	_, err := council.Trial(round, response, dest, opts, out)
	return err
}
