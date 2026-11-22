package ultrastar

import (
	"math"
	"testing"
	"time"
)

func TestBPM_IsValid(t *testing.T) {
	cases := map[string]struct {
		bpm      BPM
		expected bool
	}{
		"Zero":             {0, false},
		"Negative":         {-1, false},
		"Positive":         {120, true},
		"NaN":              {BPM(math.NaN()), false},
		"Infinity":         {BPM(math.Inf(1)), false},
		"NegativeInfinity": {BPM(math.Inf(-1)), false},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			actual := c.bpm.IsValid()
			if actual != c.expected {
				t.Errorf("BPM(%f).IsValid() = %t, expected %t", c.bpm, actual, c.expected)
			}
		})
	}
}

func TestBPM_Beats(t *testing.T) {
	cases := map[string]struct {
		bpm      BPM
		duration time.Duration
		expected Beat
	}{
		"Zero":     {120, 0, 0},
		"Regular":  {60, 2 * time.Minute, 120},
		"Negative": {-60, -30 * time.Second, 30},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			actual := c.bpm.Beats(c.duration)
			if actual != c.expected {
				t.Errorf("BPM(%f).Beats(%s) = %d, expected %d", c.bpm, c.duration, actual, c.expected)
			}
		})
	}
}

func TestBPM_Duration(t *testing.T) {
	cases := map[string]struct {
		bpm      BPM
		beats    Beat
		expected time.Duration
	}{
		"Zero":     {120, 0, 0},
		"Regular":  {60, 120, 2 * time.Minute},
		"Negative": {-60, 30, -30 * time.Second},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			actual := c.bpm.Duration(c.beats)
			if actual != c.expected {
				t.Errorf("BPM(%f).Duration(%d) = %s, expected %s", c.bpm, c.beats, actual, c.expected)
			}
		})
	}
}
