package observer

import "testing"

func sampleFrames() []Metrics {
	frames := make([]Metrics, 4)
	for i := range frames {
		frames[i] = Metrics{Tick: uint64(i * 10000), ActiveGenomes: []Genome{
			{Hash: "reproducer", Count: 20, Frequency: 0.5, Copies: uint64(i * 10), Instructions: uint64(i * 1000), Binds: uint64(i * 10), Converted: [2]int64{int64(i * 100), int64(i * 100)}},
			{Hash: "old-survivor", Count: 20, Copies: 10000},
		}}
	}
	return frames
}

func TestPersistenceRequiresNewOffspringInEveryWindow(t *testing.T) {
	frames := sampleFrames()
	r, err := Persistent(frames, PersistenceOptions{3, 5, 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Genomes) != 1 || r.Genomes[0].Copies != 30 || r.Genomes[0].Profile != "frequent-binds/few-moves/mixed-reactions" {
		t.Fatalf("wrong persistence: %+v", r)
	}
	frames[2].ActiveGenomes[0].Copies = frames[1].ActiveGenomes[0].Copies
	r, err = Persistent(frames, PersistenceOptions{3, 5, 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Genomes) != 0 {
		t.Fatal("one stagnant window was ignored")
	}
}

func TestMissingAndRareGenomesAreCensored(t *testing.T) {
	frames := sampleFrames()
	frames[1].ActiveGenomes = frames[1].ActiveGenomes[1:]
	r, err := Persistent(frames, PersistenceOptions{3, 5, 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Genomes) != 0 {
		t.Fatal("missing cumulative counter treated as zero")
	}
	frames = sampleFrames()
	frames[1].ActiveGenomes[0].Count = 1
	r, err = Persistent(frames, PersistenceOptions{3, 5, 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Genomes) != 0 {
		t.Fatal("rare intermediate population ignored")
	}
}

func TestPersistenceRejectsInvalidTelemetry(t *testing.T) {
	frames := sampleFrames()
	frames[1].Tick = 0
	if _, err := Persistent(frames, PersistenceOptions{3, 5, 5}); err == nil {
		t.Fatal("duplicate tick accepted")
	}
	frames = sampleFrames()
	frames[3].ActiveGenomes[0].Copies = 1
	if _, err := Persistent(frames, PersistenceOptions{3, 5, 5}); err == nil {
		t.Fatal("decreasing counter accepted")
	}
}
