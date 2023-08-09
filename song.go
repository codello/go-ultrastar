package ultrastar

import (
	"math"
	"time"
)

// A Song is an implementation of an UltraStar song.
// This implementation directly supports many of the known fields for songs,
// making it convenient to work with.
// Known fields are normalized to standard Go types,
// so you don't have to deal with the specifics of #GAP, #VIDEOGAP and so on.
//
// The Song type does not support parsing or serialization.
// To parse and write songs use the [github.com/Karaoke-Manager/go-ultrastar/txt] package.
type Song struct {
	// References to other files.
	AudioFileName      string
	VideoFileName      string
	CoverFileName      string
	BackgroundFileName string

	// A delay until Beat 0 of the Music.
	Gap time.Duration
	// A delay until the video starts.
	VideoGap time.Duration
	// UltraStar will jump into the song at this time.
	Start time.Duration
	// UltraStar will stop the song at this time.
	End time.Duration
	// If specified the preview will start at this time.
	PreviewStart time.Duration
	// In medley mode this is the start of the song.
	MedleyStartBeat Beat
	// If medley mode this is the end of the song.
	MedleyEndBeat Beat

	// Song metadata
	Title    string
	Artist   string
	Genre    string
	Edition  string
	Creator  string
	Language string
	Year     int
	Comment  string

	// Name of player 1
	DuetSinger1 string
	// Name of player 2
	DuetSinger2 string

	// Any custom tags that are not supported by this package.
	CustomTags map[string]string

	// Music of player 1
	MusicP1 *Music
	// Music of player 2. Any non-nil value indicates that this is a duet.
	MusicP2 *Music
}

// NewSong creates a new (single-player) song.
// Note that s.Music does not have a BPM value set.
func NewSong() (s *Song) {
	return &Song{
		CustomTags: make(map[string]string, 0),
		MusicP1:    NewMusic(),
	}
}

// NewSongWithBPM creates a new (single-player) song and
// sets the BPM of s.MusicP1 to bpm.
func NewSongWithBPM(bpm BPM) (s *Song) {
	return &Song{
		CustomTags: make(map[string]string, 0),
		MusicP1:    NewMusicWithBPM(bpm),
	}
}

// NewDuet creates a new duet.
// Note that neither s.MusicP1 nor s.MusicP2 have a BPM value set.
func NewDuet() (s *Song) {
	s = NewSong()
	s.MusicP2 = NewMusic()
	return s
}

// NewDuetWithBPM creates a new duet and
// sets the BPM of s.MusicP1 and s.MusicP2 to bpm.
func NewDuetWithBPM(bpm BPM) (s *Song) {
	s = NewSongWithBPM(bpm)
	s.MusicP2 = NewMusicWithBPM(bpm)
	return s
}

// IsDuet indicates whether a song is duet.
// Accessing s.MusicP2 is only valid for duets.
func (s *Song) IsDuet() bool {
	return s.MusicP2 != nil
}

// StartingBPM returns the BPM of s at beat 0.
func (s *Song) StartingBPM() BPM {
	if s.MusicP1 == nil {
		return BPM(math.NaN())
	}
	bpm := s.MusicP1.StartingBPM()
	if s.IsDuet() && s.MusicP2.StartingBPM() != bpm {
		return BPM(math.NaN())
	}
	return bpm
}

// BPM returns the BPM of s.
// This is intended for songs with a single BPM value, if the BPM value of s is not well-defined, NaN is returned.
// Calling this method on a song without BPM or with different BPMs for the players will return NaN.
func (s *Song) BPM() BPM {
	if s.MusicP1 == nil {
		return BPM(math.NaN())
	}
	bpm := s.MusicP1.BPM()
	if s.IsDuet() && s.MusicP2.BPM() != bpm {
		return BPM(math.NaN())
	}
	return bpm
}

// SetBPM sets the BPM of s at time 0.
// This is intended for songs with a single BPM value.
func (s *Song) SetBPM(bpm BPM) {
	if s.MusicP1 == nil {
		s.MusicP1 = NewMusic()
	}
	s.MusicP1.SetBPM(bpm)
	if s.IsDuet() {
		s.MusicP2.SetBPM(bpm)
	}
}

// Duration calculates the singing duration of s.
// The singing duration is the time from the beginning of the song until the last sung note.
func (s *Song) Duration() time.Duration {
	d := time.Duration(0)
	if s.MusicP1 != nil {
		d = s.MusicP1.Duration()
	}
	if s.IsDuet() {
		d2 := s.MusicP2.Duration()
		if d2 > d {
			d = d2
		}
	}
	d += s.Gap
	return d
}

// TODO: Function to minimize or maximize the Gap
