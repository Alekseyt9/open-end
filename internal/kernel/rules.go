package kernel

import (
	"encoding/json"
	"fmt"
	"open-end/internal/dsl"
	"open-end/internal/world"
	"sort"
)

func cloneRuleState(w *world.World) *dsl.State {
	if w.RuleState == nil {
		return &dsl.State{Usage: map[string]uint64{}}
	}
	b, _ := json.Marshal(w.RuleState)
	var s dsl.State
	_ = json.Unmarshal(b, &s)
	return &s
}

// ReloadRules is a host API. Installation is transactional, only between ticks;
// all queued future transitions are revalidated before replacing active state.
func ReloadRules(w *world.World, m *dsl.Module) error {
	if w.Config.MetabolicSwitchCost > 0 {
		return fmt.Errorf("metabolic switching currently requires baseline reactions without DSL")
	}
	if !w.Config.Ecology {
		return fmt.Errorf("DSL reactions require ecology")
	}
	if m == nil {
		return fmt.Errorf("nil rule module")
	}
	frozen, err := dsl.Compile(m.Source)
	if err != nil {
		return err
	}
	if err := m.Validate(); err != nil {
		return err
	}
	s := cloneRuleState(w)
	if len(s.History) >= 32 || len(s.Events)+len(s.Pending) >= 256 {
		return fmt.Errorf("rule history limit reached")
	}
	s.Apply(dsl.Change{Tick: w.Tick, Module: frozen})
	if err := s.Validate(w.Tick); err != nil {
		return err
	}
	w.RuleState = s
	return nil
}

func RollbackRules(w *world.World) error {
	s := cloneRuleState(w)
	if len(s.History) == 0 {
		return fmt.Errorf("no rule version to roll back")
	}
	if len(s.Events)+len(s.Pending) >= 256 {
		return fmt.Errorf("rule event limit reached")
	}
	s.Apply(dsl.Change{Tick: w.Tick, Rollback: true})
	if err := s.Validate(w.Tick); err != nil {
		return err
	}
	w.RuleState = s
	return nil
}

func ScheduleRules(w *world.World, c dsl.Change) error {
	if w.Config.MetabolicSwitchCost > 0 {
		return fmt.Errorf("metabolic switching currently requires baseline reactions without DSL")
	}
	if !w.Config.Ecology {
		return fmt.Errorf("DSL reactions require ecology")
	}
	if c.Module != nil {
		if err := c.Module.Validate(); err != nil {
			return err
		}
		m, err := dsl.Compile(c.Module.Source)
		if err != nil {
			return err
		}
		c.Module = m
	}
	s := cloneRuleState(w)
	s.Pending = append(s.Pending, c)
	sort.Slice(s.Pending, func(i, j int) bool { return s.Pending[i].Tick < s.Pending[j].Tick })
	if err := s.Validate(w.Tick); err != nil {
		return err
	}
	w.RuleState = s
	return nil
}
