package world

import "fmt"

type EnvironmentActor struct {
	Built     int64  `json:"built"`
	Reclaimed int64  `json:"reclaimed"`
	Emitted   int64  `json:"emitted"`
	Sensed    uint64 `json:"sensed"`
}
type EnvironmentState struct {
	Built              int64                        `json:"built"`
	Reclaimed          int64                        `json:"reclaimed"`
	Eroded             int64                        `json:"eroded"`
	Emitted            int64                        `json:"emitted"`
	Decayed            int64                        `json:"decayed"`
	BlockedLight       int64                        `json:"blocked_light"`
	AttenuatedTransfer int64                        `json:"attenuated_transfer"`
	Sensed             uint64                       `json:"sensed"`
	Actors             map[string]*EnvironmentActor `json:"actors,omitempty"`
}

func (w *World) EnvironmentActor(genome string) *EnvironmentActor {
	if w.Environment.Actors == nil {
		w.Environment.Actors = map[string]*EnvironmentActor{}
	}
	if w.Environment.Actors[genome] == nil {
		w.Environment.Actors[genome] = &EnvironmentActor{}
	}
	return w.Environment.Actors[genome]
}
func (w *World) ValidateEnvironment() error {
	if (w.Config.Environment != "") != (w.Environment != nil) {
		return fmt.Errorf("environment state/config mismatch")
	}
	var terrain, signal int64
	for _, c := range w.Cells {
		if c.Terrain < 0 || c.Terrain > 8 || c.Signal < 0 || c.Signal > 64 {
			return fmt.Errorf("invalid terrain or signal field")
		}
		terrain += int64(c.Terrain)
		signal += int64(c.Signal)
	}
	e := w.Environment
	if e == nil {
		if terrain != 0 || signal != 0 {
			return fmt.Errorf("environment fields in legacy world")
		}
		return nil
	}
	if e.Built < 0 || e.Reclaimed < 0 || e.Eroded < 0 || e.Emitted < 0 || e.Decayed < 0 || e.BlockedLight < 0 || e.AttenuatedTransfer < 0 || e.Reclaimed > e.Built || e.Eroded > e.Built-e.Reclaimed || e.Decayed > e.Emitted {
		return fmt.Errorf("invalid environment counters")
	}
	if e.Built-e.Reclaimed-e.Eroded != terrain || e.Emitted-e.Decayed != signal {
		return fmt.Errorf("environment fields do not reconcile")
	}
	var built, reclaimed, emitted int64
	var sensed uint64
	for hash, a := range e.Actors {
		if a == nil || w.Genomes[hash] == nil || a.Built < 0 || a.Reclaimed < 0 || a.Emitted < 0 || a.Built > e.Built-built || a.Reclaimed > e.Reclaimed-reclaimed || a.Emitted > e.Emitted-emitted || a.Sensed > e.Sensed-sensed {
			return fmt.Errorf("invalid environment actor")
		}
		built += a.Built
		reclaimed += a.Reclaimed
		emitted += a.Emitted
		sensed += a.Sensed
	}
	if built != e.Built || reclaimed != e.Reclaimed || emitted != e.Emitted || sensed != e.Sensed {
		return fmt.Errorf("environment actors do not reconcile")
	}
	return nil
}
