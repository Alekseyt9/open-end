package dsl

import (
	"strings"
	"testing"
)

func TestStateLimitsAndFutureRollbacks(t *testing.T) {
	m, err := Parse(strings.NewReader(valid))
	if err != nil {
		t.Fatal(err)
	}
	for name, s := range map[string]*State{
		"no prior version": {Pending: []Change{{Tick: 10, Rollback: true}}},
		"past change":      {Pending: []Change{{Tick: 0, Module: m}}},
		"duplicate tick":   {Pending: []Change{{Tick: 10, Module: m}, {Tick: 10, Module: m}}},
		"history overflow": {History: make([]*Module, 32), Pending: []Change{{Tick: 10, Module: m}}},
		"queue overflow":   {Pending: make([]Change, 33)},
		"event overflow":   {Events: make([]Event, 256), Pending: []Change{{Tick: 10, Module: m}}},
	} {
		t.Run(name, func(t *testing.T) {
			if s.Validate(1) == nil {
				t.Fatal("accepted invalid state")
			}
		})
	}
	s := &State{Pending: []Change{{Tick: 10, Module: m}, {Tick: 20, Rollback: true}}}
	if err := s.Validate(1); err != nil {
		t.Fatal(err)
	}
}
