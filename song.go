package ultrastar

import (
	"net/url"
	"time"
)

// Song is a data structure representing an UltraStar song. When reading a Song
// using the Reader type, some header values are promoted to struct fields and
// converted into native Go types.
type Song struct {
	Title      string
	Artist     []string
	Rendition  string
	Year       int
	Genre      []string
	Language   []string
	Edition    []string
	Tags       []string
	Creator    []string
	ProvidedBy string
	Comment    string

	// media references
	Audio           string
	AudioURL        *url.URL
	Vocals          string
	VocalsURL       *url.URL
	Instrumental    string
	InstrumentalURL *url.URL
	Video           string
	VideoURL        *url.URL
	Cover           string
	CoverURL        *url.URL
	Background      string
	BackgroundURL   *url.URL

	BPM      BPM // must not be 0
	Gap      time.Duration
	VideoGap time.Duration
	Start    time.Duration
	End      time.Duration

	// Medley and Preview
	PreviewStart time.Duration
	MedleyStart  time.Duration
	MedleyEnd    time.Duration

	// Header contains non-standard headers of the Song. Standard headers are stored
	// as struct fields instead and take precedence over the values in the Header
	// map.
	Header Header

	// Voices contain the names and notes of the song's voices. The order of the
	// voices in the slice determines the mapping of voices to the player number.
	// You can use the constants P1 through P9 to index this slice (if that many
	// voices exist in the song).
	Voices []*Voice

	_ struct{} // enforce keyed fields
}

// IsDuet indicates whether a song is a duet. Any song with more than a single
// voice is considered a duet. This method considers empty voices as well.
func (s *Song) IsDuet() bool {
	return len(s.Voices) > 1
}

// Duration calculates the singing duration of s. The singing duration is the
// time from the beginning of the song (or s.Start) until the last sung note (or
// until s.End).
func (s *Song) Duration() time.Duration {
	d := time.Duration(0)
	for _, v := range s.Voices {
		d = max(d, v.Duration(s.BPM))
	}
	d += s.Gap
	if s.End > 0 {
		d = min(d, s.End)
	}
	d -= s.Start
	return d
}

// UpdateGap sets s.Gap without changing the absolute start times of notes.
// The start of every note is adjusted to match the new gap value.
func (s *Song) UpdateGap(gap time.Duration) {
	if s.Gap == gap {
		return
	}
	delta := s.BPM.Beats(gap - s.Gap)
	for _, voice := range s.Voices {
		for i := range voice.Notes {
			voice.Notes[i].Start -= delta
		}
	}
	s.Gap = gap
}
