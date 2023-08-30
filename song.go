package ultrastar

import (
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

	// The BPM of the song.
	BPM BPM
	// A delay until Beat 0 of the song's notes.
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
	// Disable medley and preview calculation.
	NoAutoMedley bool

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
	// CustomTags are case-sensitive.
	// Note however, that the [codello.dev/ultrastar/txt] package normalizes all tags to upper case.
	CustomTags map[string]string

	// Notes of player 1.
	NotesP1 Notes
	// Notes of player 2. Any non-nil value indicates that this is a duet.
	NotesP2 Notes
}

// IsDuet indicates whether a song is duet.
// Accessing s.NotesP2 is only valid for duets.
func (s *Song) IsDuet() bool {
	return s.NotesP2 != nil
}

// Duration calculates the singing duration of s.
// The singing duration is the time from the beginning of the song until the last sung note.
func (s *Song) Duration() time.Duration {
	d := time.Duration(0)
	if s.NotesP1 != nil {
		d = s.NotesP1.Duration(s.BPM)
	}
	if s.IsDuet() {
		d2 := s.NotesP2.Duration(s.BPM)
		if d2 > d {
			d = d2
		}
	}
	d += s.Gap
	return d
}

// TODO: Function to minimize or maximize the Gap
