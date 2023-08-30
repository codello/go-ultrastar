package ultrastar

import (
	"math"
	"sort"
	"strings"
	"time"
)

// Notes represents a sequence of notes in a karaoke song.
// This usually corresponds to the notes sung by a single player.
//
// A Notes value does not know about the relative mode of UltraStar files.
// All times in a Notes value are assumed to be absolute.
// The [github.com/Karaoke-Manager/go-ultrastar/txt] package can parse and write
// UltraStar files in absolute or relative mode.
//
// All functions operating on a Notes value must maintain the invariant
// that all notes are sorted by their respective Start values.
// Notes implements [sort.Interface] which can restore this property.
type Notes []Note

// Len returns the number of notes in the slice.
//
// This is part of the implementation of [sort.Interface].
func (ns Notes) Len() int {
	return len(ns)
}

// The Less function returns a boolean value indicating whether the note at
// index i starts before note at index j.
//
// This is part of the implementation of [sort.Interface].
func (ns Notes) Less(i int, j int) bool {
	return ns[i].Start < ns[j].Start
}

// Swap swaps the notes at indexes i and j.
//
// This is part of the implementation of [sort.Interface].
func (ns Notes) Swap(i int, j int) {
	ns[i], ns[j] = ns[j], ns[i]
}

// AddNote inserts n into m.Notes white maintaining the sort property.
func AddNote(ns Notes, n Note) Notes {
	i := sort.Search(len(ns), func(i int) bool {
		return ns[i].Start > n.Start
	})
	ns = append(ns, Note{})
	copy(ns[i+1:], ns[i:])
	ns[i] = n
	return ns
}

// Duration calculates the absolute duration of m, using the specified BPM.
// The duration ignores any trailing line breaks.
func (ns Notes) Duration(bpm BPM) time.Duration {
	lastBeat := ns.LastBeat()
	return bpm.Duration(lastBeat)
}

// LastBeat calculates the last meaningful Beat in m,
// that is the last beat of the last non line break note.
func (ns Notes) LastBeat() Beat {
	for i := len(ns) - 1; i >= 0; i-- {
		if !ns[i].Type.IsLineBreak() {
			return ns[i].Start + ns[i].Duration
		}
	}
	// Either empty notes or only line breaks
	return 0
}

// ConvertToLeadingSpaces ensures that the text of notes does not end with a whitespace.
// It does so by "moving" the whitespace to the neighboring notes.
// Spaces are not moved across line breaks,
// so Notes before line breaks and the last note will have trailing spaces removed.
//
// Only the space character is understood as whitespace.
func (ns Notes) ConvertToLeadingSpaces() {
	for i := range ns[0 : len(ns)-1] {
		for strings.HasSuffix(ns[i].Text, " ") {
			ns[i].Text = ns[i].Text[0 : len(ns[i].Text)-1]
			if !ns[i+1].Type.IsLineBreak() {
				ns[i+1].Text = " " + ns[i+1].Text
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
func (ns Notes) ConvertToTrailingSpaces() {
	for i := range ns[1:] {
		for strings.HasPrefix(ns[i].Text, " ") {
			ns[i].Text = ns[i].Text[1:]
			if !ns[i-1].Type.IsLineBreak() {
				ns[i-1].Text += " "
			}
		}
	}
}

// Offset shifts all notes by the specified offset.
func (ns Notes) Offset(offset Beat) {
	// TODO: test this
	for i := range ns {
		ns[i].Start += offset
	}
}

// Substitute replaces note texts that exactly match one of the texts by the specified substitute text.
// This can be useful to replace the text of holding notes.
func (ns Notes) Substitute(substitute string, texts ...string) {
	// TODO: Test this
	textMap := make(map[string]struct{})
	for _, t := range texts {
		textMap[t] = struct{}{}
	}
	for i := range ns {
		if _, ok := textMap[ns[i].Text]; ok {
			ns[i].Text = substitute
		}
	}
}

// Scale rescales all notes, durations and BPM changes by the specified factor.
// This will increase or decrease the duration of m by factor.
// All times will be rounded to the nearest integer.
func (ns Notes) Scale(factor float64) {
	// TODO: Test this
	for i := range ns {
		ns[i].Start = Beat(math.Round(float64(ns[i].Start) * factor))
		ns[i].Duration = Beat(math.Round(float64(ns[i].Duration) * factor))
	}
}

// ScaleBPM recalculates note starts and durations to fit the specified target BPM.
// After this method returns ns.Duration(to) is approximately equal to
// ns.Duration(from) before this method was called.
// Values are rounded to the nearest integer.
func (ns Notes) ScaleBPM(from BPM, to BPM) {
	ns.Scale(float64(to / from))
}

// EnumerateLines calls f for each line of the lyrics.
// A line are the notes up to but not including a line break.
// The Start value of the following line break is passed to f as a second parameter.
// If a song does not end with a line break the [Music.LastBeat] value will be passed to f.
func (ns Notes) EnumerateLines(f func([]Note, Beat)) {
	// TODO: Test this!

	firstNoteInLine := 0
	for i, n := range ns {
		if n.Type.IsLineBreak() {
			f(ns[firstNoteInLine:i], n.Start)
			firstNoteInLine = i + 1
		}
	}
	if firstNoteInLine < len(ns) {
		f(ns[firstNoteInLine:], ns.LastBeat())
	}
}

// Lyrics generates the full lyrics of ns.
// The full lyrics is the concatenation of the individual [Note.Lyrics] values.
func (ns Notes) Lyrics() string {
	// TODO: Test this!
	var b strings.Builder
	for _, n := range ns {
		b.WriteString(n.Lyrics())
	}
	return b.String()
}
