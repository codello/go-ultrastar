package ultrastar

import (
	"errors"
	"strconv"
	"unicode/utf8"
)

// A Pitch represents the pitch of a note.
type Pitch int

// noteNames are the names of notes used for pitches. See [Pitch.NoteName] for details.
var noteNames = [12]string{"C", "C♯", "D", "E♭", "E", "F", "F♯", "G", "A♭", "A", "B♭", "B"}

// MustParsePitch works like [ParsePitch] but panics if the pitch cannot be parsed.
// This can be useful for testing or for compile-time constant pitches.
func MustParsePitch(s string) Pitch {
	p, err := ParsePitch(s)
	if err != nil {
		panic(err)
	}
	return p
}

// ParsePitch returns a new pitch based on the string representation of a pitch.
func ParsePitch(s string) (p Pitch, err error) {
	r, size := utf8.DecodeRuneInString(s)
	rest := s[size:]
	switch {
	case r >= 'A' && r <= 'B':
		p = Pitch((r-'A')*2) + 9
	case r >= 'C' && r <= 'E':
		p = Pitch((r - 'C') * 2)
	case r >= 'F' && r <= 'G':
		p = Pitch((r-'F')*2) + 5
	default:
		return 0, errors.New("invalid pitch name")
	}
	switch r, size = utf8.DecodeRuneInString(rest); r {
	case '#', '♯':
		p++
		rest = rest[size:]
	case 'b', '♭':
		p--
		rest = rest[size:]
	}
	octave := 4
	if len(rest) > 0 {
		if octave, err = strconv.Atoi(rest); err != nil {
			return 0, err
		}
	}
	p += Pitch((octave - 4) * len(noteNames))
	return p, nil
}

// NoteName returns the human-readable name of the pitch. Enharmonic equivalents
// are normalized to a fixed note name.
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
	octave := (int(p) / len(noteNames)) + 4
	if p < 0 {
		octave--
	}
	return octave
}

// String returns a human-readable string representation of the pitch.
func (p Pitch) String() string {
	return p.NoteName() + strconv.Itoa(p.Octave())
}
