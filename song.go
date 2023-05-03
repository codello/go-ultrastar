package ultrastar

import (
	"time"
)

type Song struct {
	AudioFile      string
	VideoFile      string
	CoverFile      string
	BackgroundFile string

	Gap             time.Duration
	VideoGap        time.Duration
	NotesGap        Beat
	Start           time.Duration
	End             time.Duration
	PreviewStart    time.Duration
	MedleyStartBeat Beat
	MedleyEndBeat   Beat
	CalcMedley      bool
	Resolution      int

	Title    string
	Artist   string
	Genre    string
	Edition  string
	Creator  string
	Language string
	Year     int
	Comment  string

	DuetSinger1 string
	DuetSinger2 string

	CustomTags map[string]string

	MusicP1 *Music
	MusicP2 *Music
}

func NewSong() *Song {
	return &Song{
		Resolution: 4,
		CalcMedley: true,
		CustomTags: make(map[string]string),
		MusicP1:    NewMusic(),
	}
}

func NewSongWithBPM(bpm BPM) *Song {
	return &Song{
		Resolution: 4,
		CalcMedley: true,
		CustomTags: make(map[string]string),
		MusicP1:    NewMusicWithBPM(bpm),
	}
}

func NewDuet() *Song {
	s := NewSong()
	s.MusicP2 = NewMusic()
	return s
}

func (f *Song) IsDuet() bool {
	return f.MusicP2 != nil
}
