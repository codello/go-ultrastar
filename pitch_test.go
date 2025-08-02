package ultrastar

import (
	"fmt"
	"testing"
)

func TestPitch_NoteName(t *testing.T) {
	cases := map[string]struct {
		pitch    Pitch
		expected string
	}{
		"C4":  {0, "C"},
		"C#4": {1, "C♯"},
		"B3":  {-1, "B"},
		"C5":  {12, "C"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			actual := c.pitch.NoteName()
			if actual != c.expected {
				t.Errorf("%q.NoteName() = %q, expected %q", c.pitch, actual, c.expected)
			}
		})
	}
}

func ExamplePitch_NoteName() {
	fmt.Println(NamedPitch("Gb4").NoteName())
	// Output: F♯
}

func TestPitch_Octave(t *testing.T) {
	cases := map[string]struct {
		pitch    Pitch
		expected int
	}{
		"C4":  {0, 4},
		"B4":  {11, 4},
		"C5":  {12, 5},
		"C#5": {13, 5},
		"B3":  {-1, 3},
		"C#3": {-11, 3},
		"C2":  {-12, 2},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			actual := c.pitch.Octave()
			if actual != c.expected {
				t.Errorf("%q.Octave() = %d, expected %d", c.pitch, actual, c.expected)
			}
		})
	}
}

func ExamplePitch_Octave() {
	fmt.Println(Pitch(0).Octave())
	// Output: 4
}

func TestParsePitch(t *testing.T) {
	cases := map[string]struct {
		expected    Pitch
		expectError bool
	}{
		"C4":    {0, false},
		"C#4":   {1, false},
		"C♯5":   {13, false},
		"F":     {5, false},
		"Db4":   {1, false},
		"D♭4":   {1, false},
		"A2":    {-15, false},
		"Hello": {0, true},
		"Alpha": {0, true},
	}
	for raw, c := range cases {
		t.Run(raw, func(t *testing.T) {
			actual, err := ParsePitch(raw)
			if c.expectError && err == nil {
				t.Errorf("ParsePitch(%q) did not cause an error", raw)
			} else if !c.expectError && err != nil {
				t.Errorf("ParsePitch(%q) caused an unexpected error: %s", raw, err)
			}
			if actual != c.expected {
				t.Errorf("ParsePitch(%q) = %d, expected %d", raw, actual, c.expected)
			}
		})
	}
}

func ExampleParsePitch() {
	p, _ := ParsePitch("G♭5")
	fmt.Printf("%d - %s", p, p)
	// Output: 18 - F♯5
}

func TestPitch_String(t *testing.T) {
	cases := map[string]struct {
		pitch    Pitch
		expected string
	}{
		"C4":  {0, "C4"},
		"C#4": {1, "C♯4"},
		"A2":  {-15, "A2"},
		"C#5": {13, "C♯5"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			actual := c.pitch.String()
			if actual != c.expected {
				t.Errorf("Pitch(%d).String() = %q, expected %q", c.pitch, actual, c.expected)
			}
		})
	}
}
