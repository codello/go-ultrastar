package ultrastar

import (
	"math"
	"slices"
	"strings"
	"time"
)

// These voice markers can be used to index [Song.Voices] using known player constants.
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

type Voice struct {
	Name  string
	Notes []Note

	_ struct{}
}

// AddNote inserts n into m.Notes white maintaining the sort property.
func (v *Voice) AddNote(n Note) {
	i, _ := slices.BinarySearchFunc(v.Notes, n, Note.Compare)
	v.Notes = append(v.Notes, Note{})
	copy(v.Notes[i+1:], v.Notes[i:])
	v.Notes[i] = n
}

// AppendNotes adds n at the end of n.Notes.
// This method does not ensure that the sort property is maintained.
// This method is more performant than [Voice.AddNote], especially if you are appending multiple notes.
func (v *Voice) AppendNotes(n ...Note) {
	v.Notes = append(v.Notes, n...)
}

// Sort restores the sort property of v.Notes.
// Sorting is done using a stable sorting algorithm.
func (v *Voice) Sort() {
	// FIXME: Is this a good name? Maybe RestoreSort? Or Order? Or something?
	slices.SortStableFunc(v.Notes, Note.Compare)
}

// Duration calculates the absolute duration of m, using the specified BPM.
// The duration ignores any trailing line breaks.
func (v *Voice) Duration(bpm BPM) time.Duration {
	lastBeat := v.LastBeat()
	return bpm.Duration(lastBeat)
}

// LastBeat calculates the last meaningful Beat in m,
// that is the last beat of the last non line break note.
func (v *Voice) LastBeat() Beat {
	n, ok := v.LastNote()
	if ok {
		return n.Start + n.Duration
	}
	return 0
}

// LastNote returns the last meaningful Note in ns,
// that is the last note that is not an end-of-phrase marker.
//
// If ns contains no meaningful notes, the second return value will be false.
func (v *Voice) LastNote() (Note, bool) {
	for i := len(v.Notes) - 1; i >= 0; i-- {
		if !v.Notes[i].Type.IsEndOfPhrase() {
			return v.Notes[i], true
		}
	}
	// Either empty notes or only end-of-phrase markers
	return Note{}, false
}

// IsEmpty determines if v.Notes is considered empty.
// A voice is empty if it contains no notes or only end-of-phrase markers.
//
// If v is nil, IsEmpty returns true.
func (v *Voice) IsEmpty() bool {
	if v == nil {
		return true
	}
	for _, n := range v.Notes {
		if !n.Type.IsEndOfPhrase() {
			return false
		}
	}
	return true
}

// ConvertToLeadingSpaces ensures that the text of notes does not end with a whitespace.
// It does so by "moving" the whitespace to the neighboring notes.
// Spaces are not moved across line breaks,
// so Notes before line breaks and the last note will have trailing spaces removed.
//
// Only the space character is understood as whitespace.
func (v *Voice) ConvertToLeadingSpaces() {
	for i, n := range v.Notes[0 : len(v.Notes)-1] {
		for strings.HasSuffix(n.Text, " ") {
			v.Notes[i].Text = n.Text[0 : len(n.Text)-1]
			if !n.Type.IsEndOfPhrase() {
				v.Notes[i+1].Text = " " + n.Text
			}
		}
	}
}

// ConvertToTrailingSpaces ensures that the text of notes does not start with a whitespace.
// It does so by "moving" the whitespace to the neighboring notes.
// Spaces are not moved across line breaks,
// so Notes after line breaks and the first note will have leading spaces removed.
//
// Only the space character is understood as whitespace.
func (v *Voice) ConvertToTrailingSpaces() {
	for i, n := range v.Notes[1:] {
		for strings.HasPrefix(n.Text, " ") {
			v.Notes[i].Text = n.Text[1:]
			if !n.Type.IsEndOfPhrase() {
				v.Notes[i-1].Text += " "
			}
		}
	}
}

// Offset shifts all notes by the specified offset.
func (v *Voice) Offset(offset Beat) {
	// TODO: test this
	for i := range v.Notes {
		v.Notes[i].Start += offset
	}
}

// Transpose shifts all notes in v by the given pitch.
// This corresponds to a musical transposition from the key C into the key specified by delta.
func (v *Voice) Transpose(delta Pitch) {
	// TODO: Tests
	for i := range v.Notes {
		v.Notes[i].Pitch += delta
	}
}

// Substitute replaces note texts that exactly match one of the texts by the specified substitute text.
// This can be useful to replace the text of holding notes.
func (v *Voice) Substitute(substitute string, texts ...string) {
	// TODO: Test this
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

// Scale rescales all note starts and durations by the specified factor.
// This will increase or decrease the duration of m by factor.
// All times will be rounded to the nearest integer.
func (v *Voice) Scale(factor float64) {
	// TODO: Test this
	for i, n := range v.Notes {
		v.Notes[i].Start = Beat(math.Round(float64(n.Start) * factor))
		v.Notes[i].Duration = Beat(math.Round(float64(n.Duration) * factor))
	}
}

// ScaleBPM recalculates note starts and durations to fit the specified target BPM.
// After this method returns ns.Duration(to) is approximately equal to
// ns.Duration(from) before this method was called.
// Values are rounded to the nearest integer.
func (v *Voice) ScaleBPM(from BPM, to BPM) {
	v.Scale(float64(to / from))
}

// EnumerateLines calls f for each line of the lyrics.
// A line are the notes up to but not including a line break.
// The Start value of the following line break is passed to f as a second parameter.
// If a song does not end with a line break the [Music.LastBeat] value will be passed to f.
func (v *Voice) EnumerateLines(f func([]Note, Beat)) {
	// TODO: Test this!

	firstNoteInLine := 0
	for i, n := range v.Notes {
		if n.Type.IsEndOfPhrase() {
			f(v.Notes[firstNoteInLine:i], n.Start)
			firstNoteInLine = i + 1
		}
	}
	if firstNoteInLine < len(v.Notes) {
		f(v.Notes[firstNoteInLine:], v.LastBeat())
	}
}

// Lyrics generates the full lyrics of ns.
// The full lyrics is the concatenation of the individual [Note.Lyrics] values.
func (v *Voice) Lyrics() string {
	// TODO: Test this!
	var b strings.Builder
	for _, n := range v.Notes {
		b.WriteString(n.Lyrics())
	}
	return b.String()
}

func (v *Voice) Copy() *Voice {
	return &Voice{
		Name:  v.Name,
		Notes: append([]Note(nil), v.Notes...),
	}
}
