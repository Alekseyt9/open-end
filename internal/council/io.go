// Package council exchanges evidence and proposals with a human or chat AI.
// It never calls an AI provider or executes commands contained in responses.
package council

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

var identifier = regexp.MustCompile(`^[a-z][a-z0-9-]{0,47}$`)

func digest(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
func jsonHash(v any) string  { b, _ := json.Marshal(v); return digest(b) }
func writeNew(path string, b []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr
}
func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return writeNew(path, append(b, '\n'))
}
func readJSON(path string, v any) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, 64*1024*1024+1))
	if err != nil {
		return err
	}
	if len(b) > 64*1024*1024 {
		return fmt.Errorf("JSON exceeds 64 MiB")
	}
	return decode(b, v)
}
func decode(b []byte, v any) error {
	d := json.NewDecoder(bytes.NewReader(b))
	if err := unique(d, 0); err != nil {
		return err
	}
	if _, err := d.Token(); err != io.EOF {
		return fmt.Errorf("expected one JSON document")
	}
	d = json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	return d.Decode(v)
}
func unique(d *json.Decoder, depth int) error {
	if depth > 64 {
		return fmt.Errorf("excessive JSON nesting")
	}
	t, err := d.Token()
	if err != nil {
		return err
	}
	delim, ok := t.(json.Delim)
	if !ok {
		return nil
	}
	if delim != '{' && delim != '[' {
		return fmt.Errorf("unexpected delimiter")
	}
	seen := map[string]bool{}
	for d.More() {
		if delim == '{' {
			k, err := d.Token()
			if err != nil {
				return err
			}
			s, ok := k.(string)
			if !ok || seen[s] {
				return fmt.Errorf("duplicate JSON key")
			}
			seen[s] = true
		}
		if err := unique(d, depth+1); err != nil {
			return err
		}
	}
	_, err = d.Token()
	return err
}
func leaf(s string) bool {
	return s != "" && s != "." && s != ".." && s == filepath.Base(s) && !bytes.ContainsAny([]byte(s), `/\:`)
}
