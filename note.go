package ultrastar

import (
	"bytes"
	"cmp"
	"encoding/binary"
	"fmt"
	"io"
)

// A Beat is the measurement unit for notes in a song. A beat is not an absolute
// measurement of time but must be viewed relative to the BPM value of the
// [Music].
type Beat int

// MaxBeat is the maximum value for the [Beat] type.
const MaxBeat = Beat(^uint(0) >> 1)

// The NoteType of a [Note] determines how a note is to be sung and rated.
type NoteType byte

// These are the standard note types. For details see section 4.1 of the
// UltraStar file format specification.
const (
	// NoteTypeEndOfPhrase indicates the end of a musical phrase. Usually this
	// corresponds to a line break in the lyrics of a song. End-of-phrase markers do
	// not have a duration, pitch or text.
	NoteTypeEndOfPhrase NoteType = '-'
	// NoteTypeRegular indicates a normal, sung note.
	NoteTypeRegular NoteType = ':'
	// NoteTypeGolden indicates a golden note that can award additional points.
	NoteTypeGolden NoteType = '*'
	// NoteTypeFreestyle indicates freestyle notes that are not scored.
	NoteTypeFreestyle NoteType = 'F'
	// NoteTypeRap indicates rap notes where the pitch is irrelevant.
	NoteTypeRap NoteType = 'R'
	// NoteTypeGoldenRap indicates a golden rap note that can award additional
	// points.
	NoteTypeGoldenRap NoteType = 'G'
)

// A Note represents the smallest timed unit of text in a song. Usually this
// corresponds to a syllable of text.
type Note struct {
	Type     NoteType // note type
	Start    Beat     // absolute start beat
	Duration Beat     // number of beats that the note is held
	Pitch    Pitch    // pitch of the note
	Text     string   // lyric, including whitespace
}

// String returns a string representation of the note, inspired by the UltraStar
// file format. It is not guaranteed that this method returns a string that is
// compatible with the UltraStar file format. If you need compatible
// serialization, use the Writer type.
func (n Note) String() string {
	if n.Type == NoteTypeEndOfPhrase {
		return fmt.Sprintf("%c %d", n.Type, n.Start)
	} else {
		return fmt.Sprintf("%c %d %d %d %s", n.Type, n.Start, n.Duration, n.Pitch, n.Text)
	}
}

// Lyrics returns the lyrics of the note. This is either the note's Text or may
// be a special value depending on the note type.
func (n Note) Lyrics() string {
	if n.Type == NoteTypeEndOfPhrase {
		return "\n"
	}
	return n.Text
}

// CompareStart compares n to n2 and returns an integer indicating which of the
// notes starts before the other.
func (n Note) CompareStart(n2 Note) int {
	return cmp.Compare(n.Start, n2.Start)
}

// GobEncode encodes n into a byte slice.
func (n Note) GobEncode() ([]byte, error) {
	var bs []byte
	if n.Type == NoteTypeEndOfPhrase {
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
	if n.Type == NoteTypeEndOfPhrase {
		return bs, nil
	}
	bs = binary.AppendVarint(bs, int64(n.Duration))
	bs = binary.AppendVarint(bs, int64(n.Pitch))
	bs = append(bs, []byte(n.Text)...)
	return bs, nil
}

// GobDecode decodes n from the encoded byte slice.
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
	if n.Type == NoteTypeEndOfPhrase {
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
