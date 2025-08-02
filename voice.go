package ultrastar

import (
	"iter"
	"math"
	"slices"
	"strings"
	"time"
)

// These voice markers can be used to index [Song.Voices] using known player
// constants. For example, you can use song.Voices[P1] to get the voice of
// player 1.
const (
	P1 = iota
	P2
	P3
	P4
	P5
	P6
	P7
	P8
	P9
)

// Voice represents a single voice in an UltraStar song. A Voice consists of a
// name and a sequence of notes. Note that the Voice type (any anything else in
// the [ultrastar] package) only uses absolute timestamps. The Reader and Writer
// types are able to convert from and to relative mode respectively. But notes
// in a Voice should only use absolute times.
//
// The unit of time in a Voice is a Beat. A Beat is an arbitrary quantized unit
// of time. The BPM type translates between Beat values and time.Duration.
type Voice struct {
	// Name is the name of the voice. This is usually the original singer's name and
	// will be serialized by Writer as the P1, P2, ... headers.
	Name string

	// Notes is the sequence of notes in the voice. The methods of Voice and other
	// functions in the [ultrastar] package expect the notes to be sorted in
	// ascending order by start time. Violating this invariant may produce
	// unexpected results when calling other methods on a Voice.
	Notes []Note

	_ struct{}
}

// AddNote inserts n into v.Notes white maintaining the sort invariant.
func (v *Voice) AddNote(n Note) {
	i, _ := slices.BinarySearchFunc(v.Notes, n, Note.CompareStart)
	v.Notes = append(v.Notes, Note{})
	copy(v.Notes[i+1:], v.Notes[i:])
	v.Notes[i] = n
}

// AppendNotes adds n at the end of v.Notes. This method does not ensure that
// the sort invariant is maintained, however, this method may be more performant
// than [Voice.AddNote], especially if you are appending multiple notes.
func (v *Voice) AppendNotes(n ...Note) {
	v.Notes = append(v.Notes, n...)
}

// SortNotes restores the sort invariant of v.Notes. Sorting is done using a
// stable sorting algorithm.
func (v *Voice) SortNotes() {
	slices.SortStableFunc(v.Notes, Note.CompareStart)
}

// Duration calculates the absolute duration of m, using the specified BPM. The
// duration ignores any trailing line breaks.
func (v *Voice) Duration(bpm BPM) time.Duration {
	lastBeat := v.LastBeat()
	return bpm.Duration(lastBeat)
}

// LastBeat calculates the last meaningful Beat in m, that is the last beat of
// the last non-line break note.
func (v *Voice) LastBeat() Beat {
	n, ok := v.LastNote()
	if ok {
		return n.Start + n.Duration
	}
	return 0
}

// LastNote returns the last meaningful Note in ns i.e., is the last note that
// is not an end-of-phrase marker.
//
// If ns contains no meaningful notes, the second return value will be false.
func (v *Voice) LastNote() (Note, bool) {
	for i := len(v.Notes) - 1; i >= 0; i-- {
		if v.Notes[i].Type != NoteTypeEndOfPhrase {
			return v.Notes[i], true
		}
	}
	// Either empty notes or only end-of-phrase markers
	return Note{}, false
}

// IsEmpty determines if v.Notes is considered empty. A voice is empty if it
// contains no notes or only end-of-phrase markers.
//
// If v is nil, IsEmpty returns true.
func (v *Voice) IsEmpty() bool {
	if v == nil {
		return true
	}
	for _, n := range v.Notes {
		if n.Type != NoteTypeEndOfPhrase {
			return false
		}
	}
	return true
}

// ConvertToLeadingSpaces ensures that the text of notes does not end with a
// whitespace. It does so by "moving" the whitespace to the neighboring notes.
// Spaces are not moved across line breaks, so Notes before line breaks and the
// last note will have trailing spaces removed.
//
// Only the space character is understood as whitespace.
func (v *Voice) ConvertToLeadingSpaces() {
	for i, n := range v.Notes[0 : len(v.Notes)-1] {
		for strings.HasSuffix(n.Text, " ") {
			v.Notes[i].Text = n.Text[0 : len(n.Text)-1]
			if n.Type != NoteTypeEndOfPhrase {
				v.Notes[i+1].Text = " " + n.Text
			}
		}
	}
}

// ConvertToTrailingSpaces ensures that the text of notes does not start with a
// whitespace. It does so by "moving" the whitespace to the neighboring notes.
// Spaces are not moved across line breaks, so Notes after line breaks and the
// first note will have leading spaces removed.
//
// Only the space character is understood as whitespace.
func (v *Voice) ConvertToTrailingSpaces() {
	for i, n := range v.Notes[1:] {
		for strings.HasPrefix(n.Text, " ") {
			v.Notes[i].Text = n.Text[1:]
			if n.Type != NoteTypeEndOfPhrase {
				v.Notes[i-1].Text += " "
			}
		}
	}
}

// Offset shifts all notes by the specified offset.
func (v *Voice) Offset(offset Beat) {
	for i := range v.Notes {
		v.Notes[i].Start += offset
	}
}

// Transpose shifts all notes in v by the given pitch. This corresponds to a
// musical transposition from the key C into the key specified by delta.
func (v *Voice) Transpose(delta Pitch) {
	for i := range v.Notes {
		v.Notes[i].Pitch += delta
	}
}

// Substitute replaces note texts that exactly match one of the texts by the
// specified substitute text. This can be useful to replace the text of holding
// notes.
func (v *Voice) Substitute(substitute string, texts ...string) {
	textMap := make(map[string]struct{})
	for _, t := range texts {
		textMap[t] = struct{}{}
	}
	for i, n := range v.Notes {
		if _, ok := textMap[n.Text]; ok {
			v.Notes[i].Text = substitute
		}
	}
}

// Scale rescales all note starts and durations by the specified factor. This
// will increase or decrease the duration of m by factor. Beats are rounded to
// the nearest integer.
func (v *Voice) Scale(factor float64) {
	for i, n := range v.Notes {
		v.Notes[i].Start = Beat(math.Round(float64(n.Start) * factor))
		v.Notes[i].Duration = Beat(math.Round(float64(n.Duration) * factor))
	}
}

// ScaleBPM recalculates note starts and durations to fit the specified target
// BPM. After this method returns v.Duration(to) is approximately equal to
// v.Duration(from) before this method was called. Beats are rounded to the
// nearest integer.
func (v *Voice) ScaleBPM(from BPM, to BPM) {
	v.Scale(float64(to / from))
}

// Phrases returns a sequence of phrases. Each phrase is a slice of notes up to
// but not including the end-of-phrase note. The Start value of the following
// end-of-phrase note is passed to f as a second parameter. If a song does not
// end with an end-of-phrase note, the [Voice.LastBeat] value will be the last
// Beat in the sequence.
func (v *Voice) Phrases() iter.Seq2[[]Note, Beat] {
	return func(yield func([]Note, Beat) bool) {
		firstNoteInLine := 0
		for i, n := range v.Notes {
			if n.Type == NoteTypeEndOfPhrase {
				if !yield(v.Notes[firstNoteInLine:i], n.Start) {
					return
				}
				firstNoteInLine = i + 1
			}
		}
		if firstNoteInLine < len(v.Notes) {
			yield(v.Notes[firstNoteInLine:], v.LastBeat())
		}
	}
}

// Lyrics generates the full lyrics of ns. The full lyrics is the concatenation
// of the individual [Note.Lyrics] values.
func (v *Voice) Lyrics() string {
	var b strings.Builder
	for _, n := range v.Notes {
		b.WriteString(n.Lyrics())
	}
	return b.String()
}
