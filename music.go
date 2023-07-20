package ultrastar

import (
	"bytes"
	"encoding/binary"
	"math"
	"sort"
	"strings"
	"time"
)

// BPM is a measurement of the 'speed' of a song. It counts the number of Beat's per minute.
type BPM float64

// Beats returns the number of beats in the specified duration.
// The result is rounded down to the nearest integer.
func (b BPM) Beats(d time.Duration) Beat {
	// TODO: Test this
	return Beat(float64(b) * d.Minutes())
}

// Duration returns the time it takes for bs beats to pass.
func (b BPM) Duration(bs Beat) time.Duration {
	// TODO: Test this
	return time.Duration(float64(bs) / float64(b) * float64(time.Minute))
}

// A BPMChange indicates that the BPM value of a Music changes at a certain point in time.
// BPM changes are one of the lesser known features of UltraStar songs and
// should be used with care as they are not very well known or well-supported.
//
// A BPMChange is typically used as a value type.
type BPMChange struct {
	Start Beat
	BPM   BPM
}

// GobEncode encodes c into a byte slice.
func (c *BPMChange) GobEncode() ([]byte, error) {
	// 2 bytes for Start
	// 8 bytes for BPM
	bs := make([]byte, 0, 2+8)
	bs = binary.AppendVarint(bs, int64(c.Start))
	bs = binary.AppendUvarint(bs, math.Float64bits(float64(c.BPM)))
	return bs, nil
}

// GobDecode updates c from the encoded byte slice.
func (c *BPMChange) GobDecode(bs []byte) error {
	r := bytes.NewReader(bs)
	if s, err := binary.ReadVarint(r); err != nil {
		return err
	} else {
		c.Start = Beat(s)
	}
	if b, err := binary.ReadUvarint(r); err != nil {
		return err
	} else {
		c.BPM = BPM(math.Float64frombits(b))
	}
	return nil
}

// Music is a single voice of a karaoke song.
// Naively a Music value can be viewed as a sequence of notes.
// However, Music values do support BPM changes, one of the lesser known features of UltraStar songs.
// In most cases tough, a Music value will only have a single BPM value valid for all Notes.
//
// A Music value does not know about the relative mode of UltraStar files.
// All times in a Music value are absolute.
// The [github.com/Karaoke-Manager/go-ultrastar/txt] package can parse and write
// UltraStar files in absolute or relative mode.
//
// The Notes fields of a Music value contains the sequence of notes.
// All Music methods expect the Notes and BPMs field to be sorted by their Start values.
// All custom functions that operate on Music values are expected to maintain this property.
// You can use the [Music.Sort] method to restore the sort property.
type Music struct {
	// The notes of the music. Must be kept sorted by Start values.
	Notes Notes
	// The BPM changes of the music. Must be kept sorted by Start values.
	BPMs []BPMChange
}

// NewMusic creates a new [Music] value with some default capacities for m.Notes and m.BPMs.
//
// Note that m.BPMs is empty which may break the expectation of some methods.
func NewMusic() (m *Music) {
	// We guess that songs typically have around 600 notes
	// Most songs only have 1 BPM value
	return &Music{
		Notes: make(Notes, 0, 600),
		BPMs:  make([]BPMChange, 0, 1),
	}
}

// NewMusicWithBPM creates a new [Music] value
// with a default capacity for m.Notes and a single BPM value that is valid for the entire Music.
func NewMusicWithBPM(bpm BPM) (m *Music) {
	return &Music{
		Notes: make(Notes, 0, 600),
		BPMs:  []BPMChange{{0, bpm}},
	}
}

// Clone creates a deep copy of m.
// The returned Music value has a copy of the Notes and BPMs from m.
func (m *Music) Clone() *Music {
	if m == nil {
		return nil
	}
	clone := &Music{
		Notes: make(Notes, len(m.Notes)),
		BPMs:  make([]BPMChange, len(m.BPMs)),
	}
	copy(clone.Notes, m.Notes)
	copy(clone.BPMs, m.BPMs)
	return clone
}

// AddNote inserts n into m.Notes white maintaining the sort property.
func (m *Music) AddNote(n Note) {
	i := sort.Search(len(m.Notes), func(i int) bool {
		return m.Notes[i].Start > n.Start
	})
	m.Notes = append(m.Notes, Note{})
	copy(m.Notes[i+1:], m.Notes[i:])
	m.Notes[i] = n
}

// Sort restores the sort property of m.
// After this method returns m.Notes and m.BPMs will both be sorted by their Start values.
func (m *Music) Sort() {
	sort.Sort(m.Notes)
	sort.Slice(m.BPMs, func(i, j int) bool {
		return m.BPMs[i].Start < m.BPMs[j].Start
	})
}

// BPM returns the [BPM] of m at beat 0.
// This method is intended for Music values that only have a single BPM value for the entire Music.
// On Music values without any BPM value this method panics.
func (m *Music) BPM() BPM {
	if len(m.BPMs) == 0 {
		panic("called BPM on music without BPM")
	}
	if m.BPMs[0].Start != 0 {
		return 0
	}
	return m.BPMs[0].BPM
}

// SetBPM sets the [BPM] of m at beat 0.
// This method is intended for Music values that only have a single BPM value for the entire Music.
func (m *Music) SetBPM(bpm BPM) {
	if len(m.BPMs) == 0 || m.BPMs[0].Start != 0 {
		m.BPMs = append(m.BPMs, BPMChange{})
		copy(m.BPMs[1:], m.BPMs)
		m.BPMs[0].Start = 0
	}
	m.BPMs[0].BPM = bpm
}

// Duration calculates the absolute duration of m, respecting any BPM changes.
// The duration of a song only respects beats of m.Notes.
// Any BPM changes after [Music.LastBeat] do not influence the duration.
// This method panics if it is called on a Music with no BPM value at time 0.
//
// The maximum duration of a Music value is realistically limited to about 2500h.
// Longer Music values may give inaccurate results because of floating point imprecision.
func (m *Music) Duration() time.Duration {
	if len(m.BPMs) == 0 {
		panic("called Duration on music without BPM")
	}
	if len(m.Notes) == 0 {
		return 0
	}
	if m.BPMs[0].Start != 0 {
		panic("called Duration on music without BPM at time 0")
	}
	lastBeat := m.LastBeat()

	// simple case: only one BPM for the entire song
	if len(m.BPMs) == 1 || m.BPMs[1].Start > lastBeat {
		return m.BPM().Duration(lastBeat)
	}

	// complicated case: multiple BPMs
	last := m.BPMs[0]
	duration := time.Duration(0)
	for _, current := range m.BPMs[1:] {
		if current.Start >= lastBeat {
			break
		}
		duration += last.BPM.Duration(current.Start - last.Start)
		last = current
	}
	duration += last.BPM.Duration(lastBeat - last.Start)
	return duration
}

// LastBeat calculates the last meaningful Beat in m,
// that is the last beat of the last non line break note.
func (m *Music) LastBeat() Beat {
	for i := len(m.Notes) - 1; i >= 0; i-- {
		if !m.Notes[i].Type.IsLineBreak() {
			return m.Notes[i].Start + m.Notes[i].Duration
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
func (m *Music) ConvertToLeadingSpaces() {
	for i := range m.Notes[0 : len(m.Notes)-1] {
		for strings.HasSuffix(m.Notes[i].Text, " ") {
			m.Notes[i].Text = m.Notes[i].Text[0 : len(m.Notes[i].Text)-1]
			if !m.Notes[i+1].Type.IsLineBreak() {
				m.Notes[i+1].Text = " " + m.Notes[i+1].Text
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
func (m *Music) ConvertToTrailingSpaces() {
	for i := range m.Notes[1:len(m.Notes)] {
		for strings.HasPrefix(m.Notes[i].Text, " ") {
			m.Notes[i].Text = m.Notes[i].Text[1:len(m.Notes[i].Text)]
			if !m.Notes[i-1].Type.IsLineBreak() {
				m.Notes[i-1].Text += " "
			}
		}
	}
}

// Offset shifts all notes and BPM changes by the specified offset.
func (m *Music) Offset(offset Beat) {
	// TODO: test this
	for i := range m.Notes {
		m.Notes[i].Start += offset
	}
	for i := range m.BPMs {
		m.BPMs[i].Start += offset
	}
}

// Substitute replaces note texts that exactly match one of the texts by the specified substitute text.
// This can be useful to replace the text of holding notes.
func (m *Music) Substitute(substitute string, texts ...string) {
	// TODO: Test this
	textMap := make(map[string]struct{})
	for _, t := range texts {
		textMap[t] = struct{}{}
	}
	for i := range m.Notes {
		if _, ok := textMap[m.Notes[i].Text]; ok {
			m.Notes[i].Text = substitute
		}
	}
}

// Scale rescales all notes, durations and BPM changes by the specified factor.
// This will increase or decrease the duration of m by factor.
// All times will be rounded to the nearest integer.
func (m *Music) Scale(factor float64) {
	// TODO: Test this
	for i := range m.Notes {
		m.Notes[i].Start = Beat(math.Round(float64(m.Notes[i].Start) * factor))
		m.Notes[i].Duration = Beat(math.Round(float64(m.Notes[i].Duration) * factor))
	}
	for i := range m.BPMs {
		m.BPMs[i].Start = Beat(math.Round(float64(m.BPMs[i].Start) * factor))
	}
}

// FitBPM recalculates note starts and durations to fit the specified target.
// Values are rounded to the nearest integer.
// This method tries to change the absolute timings of notes as little as possible
// while resulting in a single-BPM version of m.
func (m *Music) FitBPM(target BPM) {
	if len(m.BPMs) == 0 {
		panic("called FitBPM on Music without BPM")
	}
	currentBPM := 0
	currentStart := float64(0)
	factor := float64(target / m.BPMs[currentBPM].BPM)
	for i := range m.Notes {
		if len(m.BPMs) > currentBPM+1 && m.BPMs[currentBPM+1].Start <= m.Notes[i].Start {
			currentBPM++
			currentStart += float64(m.BPMs[currentBPM].Start-m.BPMs[currentBPM-1].Start) * factor
			factor = float64(target / m.BPMs[currentBPM].BPM)
		}
		m.Notes[i].Start = Beat(math.Round(currentStart + float64(m.Notes[i].Start-m.BPMs[currentBPM].Start)*factor))
		m.Notes[i].Duration = Beat(math.Round(float64(m.Notes[i].Duration) * factor))
	}
	// FIXME: maybe reuse existing slice?
	m.BPMs = []BPMChange{{0, target}}
}

// EnumerateLines calls f for each line of a song.
// A line are the notes up to but not including a line break.
// The Start value of the line break is passed to f as a second parameter.
// If a song does not end with a line break the [Music.LastBeat] value will be passed to f.
func (m *Music) EnumerateLines(f func([]Note, Beat)) {
	// TODO: Test this!
	if len(m.Notes) == 0 {
		return
	}

	firstNoteInLine := 0
	for i, n := range m.Notes {
		if n.Type.IsLineBreak() {
			f(m.Notes[firstNoteInLine:i], n.Start)
			firstNoteInLine = i + 1
		}
	}
	if firstNoteInLine < len(m.Notes) {
		f(m.Notes[firstNoteInLine:len(m.Notes)], m.LastBeat())
	}
}

// Lyrics generates the full lyrics of m.
// The full lyrics is the concatenation of the individual [Note.Lyrics] values.
func (m *Music) Lyrics() string {
	// TODO: Test this!
	var b strings.Builder
	for _, n := range m.Notes {
		b.WriteString(n.Lyrics())
	}
	return b.String()
}
