package observer

import (
	"encoding/json"
	"fmt"
	"io"
)

// ReadWindow retains the preceding real boundary and all later frames.
func ReadWindow(in io.Reader, window uint64) ([]Metrics, error) {
	if window == 0 {
		return nil, fmt.Errorf("window must be positive")
	}
	d := json.NewDecoder(in)
	frames := []Metrics{}
	for {
		var m Metrics
		if err := d.Decode(&m); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if len(frames) > 0 && m.Tick <= frames[len(frames)-1].Tick {
			return nil, fmt.Errorf("non-increasing telemetry tick")
		}
		frames = append(frames, m)
		cut := uint64(0)
		if m.Tick > window {
			cut = m.Tick - window
		}
		for len(frames) > 2 && frames[1].Tick <= cut {
			frames[0] = Metrics{}
			frames = frames[1:]
		}
	}
	return frames, nil
}
