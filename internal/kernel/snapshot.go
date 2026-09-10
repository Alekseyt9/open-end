package kernel

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"open-end/internal/world"
)

type snapshot struct {
	Format int          `json:"format"`
	Kernel string       `json:"kernel_version"`
	Rules  string       `json:"rule_version"`
	World  *world.World `json:"world"`
}

func Save(out io.Writer, w *world.World) error {
	if err := w.Validate(); err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(envelope(w))
}

func Load(in io.Reader) (*world.World, error) {
	var s snapshot
	d := json.NewDecoder(in)
	d.DisallowUnknownFields()
	if err := d.Decode(&s); err != nil {
		return nil, fmt.Errorf("decode snapshot: %w", err)
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("snapshot must contain one JSON object")
	}
	if (s.Format != 2 && s.Format != 3) || s.Kernel != Version || s.Rules != RuleVersion || s.World == nil || (s.Format == 2 && s.World.RuleState != nil) || (s.Format == 3 && s.World.RuleState == nil) {
		return nil, fmt.Errorf("incompatible snapshot version")
	}
	if err := s.World.Validate(); err != nil {
		return nil, fmt.Errorf("invalid snapshot: %w", err)
	}
	return s.World, nil
}

// Hash covers all state, including RNG, ancestry, counters and versions.
func Hash(w *world.World) string {
	h := sha256.New()
	_ = json.NewEncoder(h).Encode(envelope(w))
	return hex.EncodeToString(h.Sum(nil))
}

func envelope(w *world.World) snapshot {
	format := 2
	if w.RuleState != nil {
		format = 3
	}
	return snapshot{format, Version, RuleVersion, w}
}
