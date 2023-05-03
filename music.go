package ultrastar

import (
	"sort"
)

// TODO: Document that modifications to Notes and LineBreaks should be done
//  	 carefully.

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
		panic("called BPM on song without BPM")
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
