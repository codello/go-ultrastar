package ultrastar

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPitch_NoteName(t *testing.T) {
	cases := []struct {
		name     string
		pitch    Pitch
		expected string
	}{
		{"C4", 0, "C"},
		{"C#4", 1, "C#"},
		{"B3", -1, "B"},
		{"C5", 12, "C"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, c.pitch.NoteName())
		})
	}
}

func ExamplePitch_NoteName() {
	fmt.Println(NamedPitch("Gb4").NoteName())
	// Output: F#
}

func TestPitch_Octave(t *testing.T) {
	cases := []struct {
		name     string
		pitch    Pitch
		expected int
	}{
		{"C4", 0, 4},
		{"B4", 11, 4},
		{"C5", 12, 5},
		{"C#5", 13, 5},
		{"B3", -1, 3},
		{"C#3", -11, 3},
		{"C2", -12, 2},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, c.pitch.Octave())
		})
	}
}

func ExamplePitch_Octave() {
	fmt.Println(Pitch(0).Octave())
	// Output: 4
}

func TestParsePitch(t *testing.T) {
	cases := []struct {
		name     string
		s        string
		expected Pitch
	}{
		{"C4", "C4", 0},
		{"C#4", "C#4", 1},
		{"Db4", "Db4", 1},
		{"A2", "A2", -15},
		{"C#5", "C#5", 13},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			pitch, err := PitchFromString(c.s)
			assert.NoError(t, err)
			assert.Equal(t, c.expected, pitch)
		})
	}
}

func TestPitch_String(t *testing.T) {
	cases := []struct {
		name     string
		pitch    Pitch
		expected string
	}{
		{"C4", 0, "C4"},
		{"C#4", 1, "C#4"},
		{"A2", -15, "A2"},
		{"C#5", 13, "C#5"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, c.pitch.String())
		})
	}
}
