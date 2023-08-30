package ultrastar

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// A Beat is the measurement unit for notes in a song.
// A beat is not an absolute measurement of time but must be viewed relative to the BPM value of the [Music].
type Beat int

// MaxBeat is the maximum value for the [Beat] type.
const MaxBeat = Beat(^uint(0) >> 1)

// The NoteType of a [Note] specifies the input processing and rating for that
// note.
type NoteType byte

// These are the note types supported by this package.
// These correspond to the note types supported by UltraStar.
const (
	// NoteTypeLineBreak represents a line break.
	// Line Break notes do not have a Duration or Pitch.
	NoteTypeLineBreak NoteType = '-'
	// NoteTypeRegular represents a normal, sung note.
	NoteTypeRegular NoteType = ':'
	// NoteTypeGolden represents a golden note that can award additional points.
	NoteTypeGolden NoteType = '*'
	// NoteTypeFreestyle represents freestyle notes that are not graded.
	NoteTypeFreestyle NoteType = 'F'
	// NoteTypeRap represents rap notes, where the pitch is irrelevant.
	NoteTypeRap NoteType = 'R'
	// NoteTypeGoldenRap represents golden rap notes (also known as Gangsta notes)
	// that can award additional points.
	NoteTypeGoldenRap NoteType = 'G'
)

// IsValid determines if a note type is a valid UltraStar note type.
func (n NoteType) IsValid() bool {
	switch n {
	case NoteTypeLineBreak, NoteTypeRegular, NoteTypeGolden, NoteTypeFreestyle, NoteTypeRap, NoteTypeGoldenRap:
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
	case NoteTypeRap, NoteTypeGoldenRap, NoteTypeFreestyle, NoteTypeLineBreak:
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
	case NoteTypeRegular, NoteTypeGolden, NoteTypeFreestyle, NoteTypeLineBreak:
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
	case NoteTypeRegular, NoteTypeRap, NoteTypeFreestyle, NoteTypeLineBreak:
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
	case NoteTypeRegular, NoteTypeGolden, NoteTypeRap, NoteTypeGoldenRap, NoteTypeLineBreak:
		return false
	default:
		panic("invalid note type")
	}
}

// IsLineBreak determines if a note is a line break.
func (n NoteType) IsLineBreak() bool {
	switch n {
	case NoteTypeLineBreak:
		return true
	case NoteTypeRegular, NoteTypeGolden, NoteTypeRap, NoteTypeGoldenRap, NoteTypeFreestyle:
		return false
	default:
		panic("invalid note type")
	}
}

// A Note represents the smallest timed unit of text in a song.
// Usually this  corresponds to a syllable of text.
type Note struct {
	// Type denotes the kind note.
	Type NoteType
	// Start is the start beat of the note.
	Start Beat
	// Duration is the length for which the note is held.
	Duration Beat
	// Pitch is the pitch of the note.
	Pitch Pitch
	// Text is the lyric of the note.
	Text string
}

// String returns a string representation of the note, inspired by the UltraStar TXT format.
// This format should not be relied upon.
// If you need consistent serialization use the [github.com/Karaoke-Manager/go-ultrastar/txt] subpackage.
func (n Note) String() string {
	if n.Type.IsLineBreak() {
		return fmt.Sprintf("%c %d", n.Type, n.Start)
	} else {
		return fmt.Sprintf("%c %d %d %d %s", n.Type, n.Start, n.Duration, n.Pitch, n.Text)
	}
}

// Lyrics returns the lyrics of the note.
// This is either the note's Text or may be a special value depending on the note type.
func (n Note) Lyrics() string {
	if n.Type.IsLineBreak() {
		return "\n"
	}
	return n.Text
}

// GobEncode encodes n into a byte slice.
func (n Note) GobEncode() ([]byte, error) {
	var bs []byte
	if n.Type.IsLineBreak() {
		// 1 byte for Type
		// 2 bytes for Start
		bs = make([]byte, 0, 1+2)
	} else {
		// 1 byte for Type
		// 2 bytes for Start
		// 1 byte for Duration
		// 1 byte for Pitch
		// len(n.Text) bytes for n.Text
		bs = make([]byte, 0, 1+2+1+1+len(n.Text))
	}

	bs = append(bs, byte(n.Type))
	bs = binary.AppendVarint(bs, int64(n.Start))
	if n.Type.IsLineBreak() {
		return bs, nil
	}
	bs = binary.AppendVarint(bs, int64(n.Duration))
	bs = binary.AppendVarint(bs, int64(n.Pitch))
	bs = append(bs, []byte(n.Text)...)
	return bs, nil
}

// GobDecode updates n from the encoded byte slice.
func (n *Note) GobDecode(bs []byte) error {
	r := bytes.NewReader(bs)
	if t, err := r.ReadByte(); err != nil {
		return err
	} else {
		n.Type = NoteType(t)
	}
	if s, err := binary.ReadVarint(r); err != nil {
		return err
	} else {
		n.Start = Beat(s)
	}
	if n.Type.IsLineBreak() {
		n.Text = "\n"
		return nil
	}
	if d, err := binary.ReadVarint(r); err != nil {
		return err
	} else {
		n.Duration = Beat(d)
	}
	if p, err := binary.ReadVarint(r); err != nil {
		return err
	} else {
		n.Pitch = Pitch(p)
	}
	if t, err := io.ReadAll(r); err != nil {
		return err
	} else {
		n.Text = string(t)
	}
	return nil
}
