package ultrastar

import (
	"fmt"
)

// A Beat is the measurement unit for notes in a song.
type Beat int

// MaxBeat is the maximum value for the [Beat] type.
const MaxBeat = Beat(^uint(0) >> 1)

// The NoteType of a [Note] specifies the input processing and rating for that
// note.
type NoteType rune

const (
	// NoteTypeRegular represents a normal, sung note
	NoteTypeRegular NoteType = ':'
	// NoteTypeGolden represents a golden note that can award additional points
	NoteTypeGolden NoteType = '*'
	// NoteTypeFreestyle represents freestyle notes that are not graded
	NoteTypeFreestyle NoteType = 'F'
	// NoteTypeRap represents rap notes, where the pitch is irrelevant
	NoteTypeRap NoteType = 'R'
	// NoteTypeGoldenRap represents golden rap notes (also known as Gangsta notes)
	// that can award additional points.
	NoteTypeGoldenRap NoteType = 'G'
)

// IsValid determines if a note type is a valid UltraStar note type.
func (n NoteType) IsValid() bool {
	switch n {
	case NoteTypeRegular, NoteTypeGolden, NoteTypeFreestyle, NoteTypeRap, NoteTypeGoldenRap:
		return true
	default:
		return false
	}
}

// IsSung determines if a note is a normally sung note (golden or not).
func (n NoteType) IsSung() bool {
	switch n {
	case NoteTypeRegular, NoteTypeGolden:
		return true
	case NoteTypeRap, NoteTypeGoldenRap, NoteTypeFreestyle:
		return false
	default:
		panic("invalid note type")
	}
}

// IsRap determines if a note is a rap note (golden or not).
func (n NoteType) IsRap() bool {
	switch n {
	case NoteTypeRap, NoteTypeGoldenRap:
		return true
	case NoteTypeRegular, NoteTypeGolden, NoteTypeFreestyle:
		return false
	default:
		panic("invalid note type")
	}
}

// IsGolden determines if a note is a golden note (rap or regular).
func (n NoteType) IsGolden() bool {
	switch n {
	case NoteTypeGolden, NoteTypeGoldenRap:
		return true
	case NoteTypeRegular, NoteTypeRap, NoteTypeFreestyle:
		return false
	default:
		panic("invalid note type")
	}
}

// IsFreestyle determines if a note is a freestyle note.
func (n NoteType) IsFreestyle() bool {
	switch n {
	case NoteTypeFreestyle:
		return true
	case NoteTypeRegular, NoteTypeGolden, NoteTypeRap, NoteTypeGoldenRap:
		return false
	default:
		panic("invalid note type")
	}
}

// A Note represents the smallest timed unit of text in a song. Usually this
// corresponds to a syllable of text.
type Note struct {
	// Type denotes the kind note. See [NoteType] for details.
	Type NoteType
	// Start is the start beat of the note.
	Start Beat
	// Duration is the length for which the note is held.
	Duration Beat
	// Pitch is the pitch of the note. See [Pitch] for details.
	Pitch Pitch
	// Text is the lyric of the note.
	Text string
}

// String returns a string representation of the note, inspired by the UltraStar
// TXT format.
func (n Note) String() string {
	return fmt.Sprintf("%c %d %d %d %s", n.Type, n.Start, n.Duration, n.Pitch, n.Text)
}

// Notes is an alias type for a slice of notes. This type implements the sort
// interface.
type Notes []Note

// Len returns the number of notes in the slice.
func (n Notes) Len() int {
	return len(n)
}

// The Less function returns a boolean value indicating whether the note at
// index i starts before note at index j.
func (n Notes) Less(i int, j int) bool {
	return n[i].Start < n[j].Start
}

// Swap swaps the notes at indexes i and j.
func (n Notes) Swap(i int, j int) {
	n[i], n[j] = n[j], n[i]
}
