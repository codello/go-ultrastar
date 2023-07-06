package ultrastar

import (
	"errors"
	"strconv"
)

// These known errors might be returned by some of the functions and methods in this package.
var (
	// ErrInvalidPitchName denotes that the named pitch was not recognized.
	ErrInvalidPitchName = errors.New("unknown pitch name")
)

// A Pitch represents the pitch of a note.
type Pitch int

// noteNames are the names of notes used for pitches. See [Pitch.NoteName] for details.
var noteNames = [12]string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

// NamedPitch works like [PitchFromString] but panics if the pitch cannot be parsed.
// This can be useful for testing or for compile-time constant pitches.
func NamedPitch(s string) Pitch {
	p, err := PitchFromString(s)
	if err != nil {
		panic(err)
	}
	return p
}

// PitchFromString returns a new pitch based on the string representation of a pitch.
func PitchFromString(s string) (p Pitch, err error) {
	ok := false
	for index, note := range noteNames {
		if note == string(s[0]) {
			p = Pitch(index)
			ok = true
		}
	}
	if !ok {
		return p, ErrInvalidPitchName
	}
	var rest string
	switch s[1] {
	case '#':
		p += 1
		rest = s[2:]
	case 'b':
		p -= 1
		rest = s[2:]
	default:
		rest = s[1:]
	}
	octave, err := strconv.Atoi(rest)
	if err != nil {
		octave = 4
	}
	p = Pitch(int(p) + (octave-4)*len(noteNames))
	return p, nil
}

// NoteName returns the human-readable name of the pitch.
// The note naming is not very sophisticated.
// Only whole and half steps are supported and note names use sharps exclusively.
// So a D flat and a C sharp will both return "C#" as their note name.
func (p Pitch) NoteName() string {
	i := int(p) % len(noteNames)
	if i < 0 {
		i += len(noteNames)
	}
	return noteNames[i]
}

// Octave returns the [scientific octave] of a pitch.
//
// [scientific octave]: https://en.wikipedia.org/wiki/Octave#Notation
func (p Pitch) Octave() int {
	// FIXME: Is 0 actually C4?
	octave := (int(p) / len(noteNames)) + 4
	if p < 0 {
		octave -= 1
	}
	return octave
}

// String returns a human-readable string representation of the pitch.
func (p Pitch) String() string {
	return p.NoteName() + strconv.Itoa(p.Octave())
}
