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
	Audio        []string
	Vocals       []string
	Instrumental []string
	Video        []string
	Cover        []string
	Background   []string

	BPM      BPM
	Gap      time.Duration
	VideoGap time.Duration
	Start    time.Duration
	End      time.Duration

	// Medley and Preview
	NoAutoMedley bool
	PreviewStart time.Duration
	MedleyStart  time.Duration
	MedleyEnd    time.Duration

	// Song metadata
	Title     string
	Artist    string
	Genres    []string
	Editions  []string
	Creators  []string
	Languages []string
	Year      int
	Comment   string

	// Application-specific headers, fields take precedence
	ExtraHeaders Header

	// Notes of all the voices in the song
	Voices []*Voice

	_ struct{}
}

func (s *Song) Copy() *Song {
	voices := make([]*Voice, len(s.Voices))
	for i, voice := range s.Voices {
		voices[i] = voice.Copy()
	}
	return &Song{
		Audio:        append([]string(nil), s.Audio...),
		Vocals:       append([]string(nil), s.Vocals...),
		Instrumental: append([]string(nil), s.Instrumental...),
		Video:        append([]string(nil), s.Video...),
		Cover:        append([]string(nil), s.Cover...),
		Background:   append([]string(nil), s.Background...),
		BPM:          s.BPM,
		Gap:          s.Gap,
		VideoGap:     s.VideoGap,
		Start:        s.Start,
		End:          s.End,
		NoAutoMedley: s.NoAutoMedley,
		PreviewStart: s.PreviewStart,
		MedleyStart:  s.MedleyStart,
		MedleyEnd:    s.MedleyEnd,
		Title:        s.Title,
		Artist:       s.Artist,
		Genres:       append([]string(nil), s.Genres...),
		Editions:     append([]string(nil), s.Editions...),
		Creators:     append([]string(nil), s.Creators...),
		Languages:    append([]string(nil), s.Languages...),
		Year:         s.Year,
		Comment:      s.Comment,
		ExtraHeaders: s.ExtraHeaders.Copy(),
		Voices:       voices,
	}
}

// IsDuet indicates whether a song is duet.
func (s *Song) IsDuet() bool {
	return len(s.Voices) > 1
}

// TODO: Normalize or something that performs normalization on a song
// - Remove empty voices
// - Upper case header fields
// - Trim known metadata
// - Set gap so that at least one voice starts at 0
// - Detect known headers and set the corresponding fields (if empty) or join (if multi-valued)

// Duration calculates the singing duration of s.
// The singing duration is the time from the beginning of the song until the last sung note.
func (s *Song) Duration() time.Duration {
	d := time.Duration(0)
	for _, v := range s.Voices {
		d = max(d, v.Duration(s.BPM))
	}
	d += s.Gap
	return d
}

// TODO: Function to minimize or maximize the Gap
