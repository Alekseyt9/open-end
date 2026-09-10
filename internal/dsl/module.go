// Package dsl compiles a bounded, declarative reaction language. It has no
// access to the filesystem, network, world pointers or native code execution.
package dsl

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"sort"
)

const MaxSourceBytes = 65536
const MaxRules = 16
const ResourceCount = 5

var resourceNames = [ResourceCount]string{"X", "Y", "Z", "field_energy", "energy"}
var potential = [ResourceCount]int{8, 4, 0, 1, 1}
var identifier = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{0,47}$`)

type Document struct {
	Format            int        `json:"format"`
	Version           string     `json:"version"`
	InstructionBudget int        `json:"instruction_budget"`
	Rules             []RuleSpec `json:"rules"`
}
type RuleSpec struct {
	ID         int            `json:"id"`
	Name       string         `json:"name"`
	Consume    map[string]int `json:"consume"`
	Produce    map[string]int `json:"produce"`
	EnergyCost int            `json:"energy_cost"`
	MaxBatch   int            `json:"max_batch"`
}
type Instruction struct {
	Op       uint8 `json:"op"`
	Resource int   `json:"resource"`
	Amount   int   `json:"amount"`
}

const (
	Consume uint8 = iota
	Produce
)

type Rule struct {
	ID         int           `json:"id"`
	Name       string        `json:"name"`
	EnergyCost int           `json:"energy_cost"`
	MaxBatch   int           `json:"max_batch"`
	Code       []Instruction `json:"code"`
}
type Module struct {
	Source Document `json:"source"`
	Hash   string   `json:"sha256"`
	Rules  []Rule   `json:"bytecode"`
}

func Parse(in io.Reader) (*Module, error) {
	b, err := io.ReadAll(io.LimitReader(in, MaxSourceBytes+1))
	if err != nil {
		return nil, err
	}
	if len(b) > MaxSourceBytes {
		return nil, fmt.Errorf("DSL exceeds 64 KiB")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	if err := uniqueValue(d, 0); err != nil {
		return nil, err
	}
	if _, err := d.Token(); err != io.EOF {
		return nil, fmt.Errorf("DSL must contain one document")
	}
	d = json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	var doc Document
	if err := d.Decode(&doc); err != nil {
		return nil, fmt.Errorf("DSL parse: %w", err)
	}
	return Compile(doc)
}

// uniqueValue rejects duplicate JSON keys and excessive nesting before decoding
// the schema. Otherwise JSON's last-key-wins behavior could conceal mistakes.
func uniqueValue(d *json.Decoder, depth int) error {
	if depth > 12 {
		return fmt.Errorf("DSL nesting exceeds 12")
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
	keys := map[string]bool{}
	for d.More() {
		if delim == '{' {
			k, err := d.Token()
			if err != nil {
				return err
			}
			name, ok := k.(string)
			if !ok || keys[name] {
				return fmt.Errorf("duplicate or invalid DSL key %v", k)
			}
			keys[name] = true
		}
		if err := uniqueValue(d, depth+1); err != nil {
			return err
		}
	}
	_, err = d.Token()
	return err
}

func Compile(input Document) (*Module, error) {
	// Freeze caller-owned maps/slices before hashing or installing.
	b, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	if len(b) > MaxSourceBytes {
		return nil, fmt.Errorf("DSL exceeds 64 KiB")
	}
	var doc Document
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, err
	}
	if doc.Format != 1 || !identifier.MatchString(doc.Version) || doc.InstructionBudget < 1 || doc.InstructionBudget > 64 || len(doc.Rules) == 0 || len(doc.Rules) > MaxRules {
		return nil, fmt.Errorf("invalid DSL format, version, budget or rule count")
	}
	sort.Slice(doc.Rules, func(i, j int) bool { return doc.Rules[i].ID < doc.Rules[j].ID })
	m := &Module{Source: doc, Rules: make([]Rule, 0, len(doc.Rules))}
	ids, names := map[int]bool{}, map[string]bool{}
	for _, s := range doc.Rules {
		if s.ID < 0 || s.ID >= MaxRules || ids[s.ID] || !identifier.MatchString(s.Name) || names[s.Name] || s.EnergyCost < 1 || s.EnergyCost > 64 || s.MaxBatch < 1 || s.MaxBatch > 64 {
			return nil, fmt.Errorf("invalid or duplicate rule %q", s.Name)
		}
		ids[s.ID] = true
		names[s.Name] = true
		r := Rule{ID: s.ID, Name: s.Name, EnergyCost: s.EnergyCost, MaxBatch: s.MaxBatch, Code: []Instruction{}}
		var consumed, produced [ResourceCount]int
		for side, values := range []map[string]int{s.Consume, s.Produce} {
			if len(values) == 0 {
				return nil, fmt.Errorf("rule %s has an empty side", s.Name)
			}
			for name, n := range values {
				known := false
				for _, valid := range resourceNames {
					known = known || name == valid
				}
				if !known || n < 1 || n > 64 {
					return nil, fmt.Errorf("unknown resource or invalid amount in %s", s.Name)
				}
			}
			for resource, name := range resourceNames {
				if n := values[name]; n > 0 {
					r.Code = append(r.Code, Instruction{Op: uint8(side), Resource: resource, Amount: n})
					if side == 0 {
						consumed[resource] = n
					} else {
						produced[resource] = n
					}
				}
			}
		}
		massIn, massOut, energy := 0, 0, 0
		for i := 0; i < ResourceCount; i++ {
			if i < 3 {
				massIn += consumed[i]
				massOut += produced[i]
			}
			energy += (produced[i] - consumed[i]) * potential[i]
		}
		if massIn == 0 || massIn != massOut || energy != 0 {
			return nil, fmt.Errorf("rule %s violates chemical matter or energy conservation", s.Name)
		}
		if consumed == produced {
			return nil, fmt.Errorf("rule %s has no effect", s.Name)
		}
		if Work(r) > doc.InstructionBudget {
			return nil, fmt.Errorf("rule %s exceeds instruction budget", s.Name)
		}
		m.Rules = append(m.Rules, r)
	}
	b, _ = json.Marshal(doc)
	hash := sha256.Sum256(b)
	m.Hash = hex.EncodeToString(hash[:])
	return m, nil
}

func (m *Module) Validate() error {
	if m == nil {
		return nil
	}
	compiled, err := Compile(m.Source)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(m, compiled) {
		return fmt.Errorf("compiled DSL or hash differs from source")
	}
	return nil
}
func (m *Module) Find(id int) *Rule {
	if m == nil {
		return nil
	}
	for i := range m.Rules {
		if m.Rules[i].ID == id {
			return &m.Rules[i]
		}
	}
	return nil
}
func Work(r Rule) int { return 2*len(r.Code) + ResourceCount }

type Effect struct {
	Units int
	Delta [ResourceCount]int
	Work  int
}

// Execute preflights all inputs and capacities before returning a transaction.
// The caller applies the returned delta only after this function succeeds.
func Execute(r Rule, available, capacity [ResourceCount]int, requested, budget int) (Effect, error) {
	if Work(r) > budget {
		return Effect{}, fmt.Errorf("instruction budget exhausted")
	}
	e := Effect{Work: Work(r)}
	units := max(0, min(requested, r.MaxBatch))
	var delta [ResourceCount]int
	for _, i := range r.Code {
		if i.Resource < 0 || i.Resource >= ResourceCount || i.Amount < 1 || i.Amount > 64 || i.Op > Produce {
			return Effect{}, fmt.Errorf("invalid bytecode")
		}
		if i.Op == Consume {
			units = min(units, available[i.Resource]/i.Amount)
			delta[i.Resource] -= i.Amount
		} else {
			delta[i.Resource] += i.Amount
		}
	}
	for i, n := range delta {
		if n > 0 {
			units = min(units, (capacity[i]-available[i])/n)
		}
	}
	units = max(0, units)
	for i, n := range delta {
		e.Delta[i] = n * units
	}
	e.Units = units
	return e, nil
}
