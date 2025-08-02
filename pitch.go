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

// NamedPitch works like [ParsePitch] but panics if the pitch cannot be parsed.
// This can be useful for testing or for compile-time constant pitches.
func NamedPitch(s string) Pitch {
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
	if r < 65 || r > 71 {
		return 0, errors.New("invalid pitch name")
	}
	if r <= 66 {
		// A, or B
		p = Pitch((r-65)*2) + 9
	} else if r <= 69 {
		// C, D, or E
		p = Pitch((r - 67) * 2)
	} else {
		// F, G
		p = Pitch((r-70)*2) + 5
	}
	switch r, size = utf8.DecodeRuneInString(rest); r {
	case '#', '♯':
		p += 1
		rest = rest[size:]
	case 'b', '♭':
		p -= 1
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
		octave -= 1
	}
	return octave
}

// String returns a human-readable string representation of the pitch.
func (p Pitch) String() string {
	return p.NoteName() + strconv.Itoa(p.Octave())
}
