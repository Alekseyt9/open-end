package dsl

import "fmt"

type Change struct {
	Tick     uint64  `json:"tick"`
	Module   *Module `json:"module,omitempty"`
	Rollback bool    `json:"rollback,omitempty"`
}
type Event struct {
	Tick   uint64 `json:"tick"`
	Action string `json:"action"`
	Hash   string `json:"sha256"`
}
type State struct {
	Active       *Module           `json:"active"`
	History      []*Module         `json:"history"`
	Pending      []Change          `json:"pending"`
	Events       []Event           `json:"events"`
	Usage        map[string]uint64 `json:"usage"`
	Instructions uint64            `json:"instructions"`
}

func (s *State) Validate(tick uint64) error {
	if s == nil {
		return nil
	}
	if len(s.History) > 32 || len(s.Pending) > 32 || len(s.Events)+len(s.Pending) > 256 {
		return fmt.Errorf("DSL revision limits exceeded")
	}
	if err := s.Active.Validate(); err != nil {
		return err
	}
	for _, m := range s.History {
		if err := m.Validate(); err != nil {
			return err
		}
	}
	depth := len(s.History)
	for i, c := range s.Pending {
		if c.Tick < tick || i > 0 && c.Tick <= s.Pending[i-1].Tick {
			return fmt.Errorf("rule changes must have unique future ticks")
		}
		if c.Rollback {
			if c.Module != nil || depth == 0 {
				return fmt.Errorf("rollback has no prior module")
			}
			depth--
		} else {
			if c.Module == nil {
				return fmt.Errorf("missing scheduled module")
			}
			if err := c.Module.Validate(); err != nil {
				return err
			}
			depth++
		}
		if depth > 32 {
			return fmt.Errorf("scheduled rule history exceeds limit")
		}
	}
	for i, e := range s.Events {
		if e.Tick > tick || i > 0 && e.Tick < s.Events[i-1].Tick || e.Action != "load" && e.Action != "rollback" {
			return fmt.Errorf("invalid rule event")
		}
	}
	return nil
}

// Apply is called at a tick boundary after the whole change sequence has been
// validated. Module bytecode is already frozen and contains no external links.
func (s *State) Apply(c Change) {
	if c.Rollback {
		s.Active = s.History[len(s.History)-1]
		s.History = s.History[:len(s.History)-1]
	} else {
		s.History = append(s.History, s.Active)
		s.Active = c.Module
	}
	hash := "builtin"
	if s.Active != nil {
		hash = s.Active.Hash
	}
	action := "load"
	if c.Rollback {
		action = "rollback"
	}
	s.Events = append(s.Events, Event{c.Tick, action, hash})
}
