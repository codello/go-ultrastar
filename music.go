package ultrastar

import (
	"sort"
	"strings"
	"time"
)

// TODO: Document that modifications to Notes and LineBreaks should be done
//  	 carefully.

// TODO: Doc maximum music duration is about 2500 hours. For longer music some
// 		 calculations may produce wrong results because of floating point
//		 precision.

// BPM is a measurement of the 'speed' of a song. It counts the number of Beat's
// per minute.
type BPM float64

type BPMChange struct {
	Start Beat
	BPM   BPM
}

type Music struct {
	Notes      Notes
	LineBreaks []Beat
	BPMs       []BPMChange
}

func NewMusic() *Music {
	// We guess some capacities for notes and line breaks
	return &Music{
		Notes:      make(Notes, 0, 400),
		LineBreaks: make([]Beat, 0, 50),
		BPMs:       make([]BPMChange, 0, 1),
	}
}

func NewMusicWithBPM(bpm BPM) *Music {
	return &Music{
		Notes:      make(Notes, 0, 400),
		LineBreaks: make([]Beat, 0, 50),
		BPMs:       []BPMChange{{0, bpm}},
	}
}

func (m *Music) AddNote(n Note) {
	i := sort.Search(len(m.Notes), func(i int) bool {
		return m.Notes[i].Start > n.Start
	})
	m.Notes = append(m.Notes, Note{})
	copy(m.Notes[i+1:], m.Notes[i:])
	m.Notes[i] = n
}

func (m *Music) Sort() {
	sort.Sort(m.Notes)
	sort.Slice(m.LineBreaks, func(i, j int) bool {
		return m.LineBreaks[i] < m.LineBreaks[j]
	})
	sort.Slice(m.BPMs, func(i, j int) bool {
		return m.BPMs[i].Start < m.BPMs[j].Start
	})
}

func (m *Music) BPM() BPM {
	if len(m.BPMs) == 0 {
		panic("called BPM on music without BPM")
	}
	if m.BPMs[0].Start != 0 {
		return 0
	}
	return m.BPMs[0].BPM
}

func (m *Music) SetBPM(bpm BPM) {
	if len(m.BPMs) == 0 || m.BPMs[0].Start != 0 {
		m.BPMs = append(m.BPMs, BPMChange{})
		copy(m.BPMs[1:], m.BPMs)
		m.BPMs[0].Start = 0
	}
	m.BPMs[0].BPM = bpm
}

func (m *Music) Duration() time.Duration {
	if len(m.BPMs) == 0 {
		panic("called Duration on music without BPM")
	}
	if len(m.Notes) == 0 {
		return 0
	}
	if m.Notes[0].Start < m.BPMs[0].Start {
		panic("called Duration on music with notes before first BPM")
	}
	lastNote := m.Notes[len(m.Notes)-1]
	lastBeat := lastNote.Start + lastNote.Duration

	// simple case: only one BPM for the entire song
	if len(m.BPMs) == 1 || m.BPMs[1].Start > lastBeat {
		return time.Duration(float64(lastBeat) / float64(m.BPM()) * float64(time.Minute))
	}

	// complicated case: multiple BPMs
	last := m.BPMs[0]
	duration := time.Duration(0)
	for _, current := range m.BPMs[1:] {
		if current.Start >= lastBeat {
			break
		}
		d := current.Start - last.Start
		duration += time.Duration(float64(d) / float64(last.BPM) * float64(time.Minute))
		last = current
	}
	d := lastBeat - last.Start
	duration += time.Duration(float64(d) / float64(last.BPM) * float64(time.Minute))
	return duration
}

func (m *Music) LastBeat() Beat {
	// TODO: Doc: Returns beat of last note, even if there are line breaks or BPM changes later
	if len(m.Notes) == 0 {
		return 0
	}
	n := m.Notes[len(m.Notes)-1]
	return n.Start + n.Duration
}

func (m *Music) ConvertToLeadingSpaces() {
	for i := range m.Notes[0 : len(m.Notes)-1] {
		// FIXME: Maybe do not move space across line breaks
		for strings.HasSuffix(m.Notes[i].Text, " ") {
			m.Notes[i].Text = m.Notes[i].Text[0 : len(m.Notes[i].Text)-1]
			m.Notes[i+1].Text = " " + m.Notes[i+1].Text
		}
	}
}

func (m *Music) ConvertToTrailingSpaces() {
	for i := range m.Notes[1:len(m.Notes)] {
		// FIXME: Maybe do not move space across line breaks
		for strings.HasPrefix(m.Notes[i].Text, " ") {
			m.Notes[i].Text = m.Notes[i].Text[1:len(m.Notes[i].Text)]
			m.Notes[i-1].Text = m.Notes[i-1].Text + " "
		}
	}
}

func (m *Music) EnumerateLines(f func([]Note, Beat)) {
	// TODO: Test this!
	if len(m.Notes) == 0 {
		return
	}

	nextBreak := 0
	firstNoteInLine := 0
	for i, n := range m.Notes {
		// No more line breaks. One single line remaining
		if nextBreak >= len(m.LineBreaks) {
			f(m.Notes[firstNoteInLine:], m.LastBeat())
			return
		}
		if m.LineBreaks[nextBreak] <= n.Start {
			f(m.Notes[firstNoteInLine:i], m.LineBreaks[nextBreak])
			nextBreak++
			firstNoteInLine = i
		}
	}
	f(m.Notes[firstNoteInLine:len(m.Notes)], m.LastBeat())
}

func (m *Music) Lyrics() string {
	// TODO: Test this!
	var b strings.Builder
	m.EnumerateLines(func(notes []Note, _ Beat) {
		for _, n := range notes {
			b.WriteString(n.Text)
		}
		b.WriteRune('\n')
	})
	return b.String()
}

// TODO: Functions:
//  	 - Iterating over lines / splitting into lines
//	     - Get Lyrics
//       - Convert Holding Notes from - to ~ and back
//       - Lengthen / Shorten Music
//       - Offset music
