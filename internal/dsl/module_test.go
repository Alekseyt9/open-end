package dsl

import (
	"strings"
	"testing"
)

const valid = `{"format":1,"version":"test-v1","instruction_budget":32,"rules":[{"id":0,"name":"x-to-y","consume":{"X":1},"produce":{"Y":1,"energy":4},"energy_cost":1,"max_batch":64}]}`

func TestParserRejectsUnsafeOrAmbiguousRules(t *testing.T) {
	for name, source := range map[string]string{
		"energy creation":  strings.Replace(valid, `"energy":4`, `"energy":5`, 1),
		"matter creation":  strings.Replace(valid, `"Y":1`, `"Y":2`, 1),
		"unknown resource": strings.Replace(valid, `"Y":1`, `"W":1`, 1),
		"negative":         strings.Replace(valid, `"X":1`, `"X":-1`, 1),
		"over budget":      strings.Replace(valid, `"instruction_budget":32`, `"instruction_budget":1`, 1),
		"unbounded budget": strings.Replace(valid, `"instruction_budget":32`, `"instruction_budget":1000`, 1),
		"unknown field":    strings.Replace(valid, `"format":1`, `"format":1,"exec":"shell"`, 1),
		"duplicate key":    strings.Replace(valid, `"X":1`, `"X":9,"X":1`, 1),
		"trailing":         valid + `{}`,
		"oversize":         valid + strings.Repeat(" ", MaxSourceBytes),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse(strings.NewReader(source)); err == nil {
				t.Fatal("accepted invalid source")
			}
		})
	}
	m, err := Parse(strings.NewReader(valid))
	if err != nil {
		t.Fatal(err)
	}
	m.Source.Rules = append(m.Source.Rules, m.Source.Rules[0])
	if _, err := Compile(m.Source); err == nil {
		t.Fatal("accepted duplicate rule")
	}
}

func TestExecuteAtomicCapacityAndBudget(t *testing.T) {
	m, err := Parse(strings.NewReader(valid))
	if err != nil {
		t.Fatal(err)
	}
	r := m.Rules[0]
	available := [ResourceCount]int{10, 0, 0, 0, 250}
	capacity := [ResourceCount]int{100, 100, 100, 128, 256}
	e, err := Execute(r, available, capacity, 64, 32)
	if err != nil || e.Units != 1 || e.Delta != [ResourceCount]int{-1, 1, 0, 0, 4} {
		t.Fatalf("capacity limit: %+v %v", e, err)
	}
	balance := 0
	for i, n := range e.Delta {
		balance += n * potential[i]
	}
	if balance != 0 || e.Delta[0]+e.Delta[1]+e.Delta[2] != 0 {
		t.Fatal("not conserved")
	}
	e, err = Execute(r, available, capacity, 64, 1)
	if err == nil || e != (Effect{}) {
		t.Fatal("budget failure returned partial effect")
	}
	available[0] = 0
	e, err = Execute(r, available, capacity, 64, 32)
	if err != nil || e.Units != 0 || e.Delta != [ResourceCount]int{} {
		t.Fatal("insufficient input changed resources")
	}
	m.Rules[0].Code[0].Amount = 2
	if m.Validate() == nil {
		t.Fatal("accepted tampered bytecode")
	}
}
