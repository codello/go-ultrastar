package ultrastar

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestPitchNoteName tests the [Pitch.NoteName] method.
func TestPitch_NoteName(t *testing.T) {
	tests := map[Pitch]string{
		0:  "C",
		1:  "C#",
		-1: "B",
		12: "C",
	}
	for pitch, expected := range tests {
		assert.Equal(t, expected, pitch.NoteName())
	}
}

// TestPitchOctave tests the [Pitch.Octave] method.
func TestPitch_Octave(t *testing.T) {
	pitches := map[Pitch]int{
		0:   4,
		11:  4,
		12:  5,
		13:  5,
		-1:  3,
		-11: 3,
		-12: 2,
	}
	for pitch, expected := range pitches {
		assert.Equal(t, expected, pitch.Octave())
	}
}

func TestParsePitch(t *testing.T) {
	results := map[string]Pitch{
		"C4":  0,
		"C#4": 1,
		"Db4": 1,
		"A2":  -15,
		"C#5": 13,
	}
	for s, expected := range results {
		pitch, err := PitchFromString(s)
		assert.NoError(t, err, s)
		assert.Equal(t, expected, pitch, s)
	}
}

// TestPitchString tests the [Pitch.String] function.
func TestPitch_String(t *testing.T) {
	pitches := map[Pitch]string{
		0:   "C4",
		1:   "C#4",
		-15: "A2",
		13:  "C#5",
	}
	for pitch, expected := range pitches {
		assert.Equal(t, expected, pitch.String())
	}
}
