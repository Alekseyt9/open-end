package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestTreeCLIHelpAndInvalidArguments(t *testing.T) {
	for _, args := range [][]string{{"tree"}, {"tree", "help"}, {"tree", "grow", "-h"}, {"help"}} {
		var out bytes.Buffer
		if err := run(args, &out); err != nil || !strings.Contains(out.String(), "tree") {
			t.Fatal(args, out.String(), err)
		}
	}
	for _, args := range [][]string{
		{"tree", "unknown"}, {"tree", "init"}, {"tree", "show"}, {"tree", "select"}, {"tree", "grow"}, {"tree", "export"},
		{"tree", "grow", "-tree", "missing", "unexpected"}, {"tree", "grow", "-workers", "oops"}, {"tree", "init", "-window", "0"},
	} {
		var out bytes.Buffer
		if err := run(args, &out); err == nil {
			t.Fatal("accepted invalid command", args)
		}
	}
}
