package ultrastar

import (
	"testing"
	"time"
)

func TestSong_Duration(t *testing.T) {
	tests := map[string]struct {
		song *Song
		want time.Duration
	}{
		"single voice": {&Song{
			BPM: 120,
			Voices: []*Voice{{Notes: []Note{
				{NoteTypeRegular, 25, 30, 0, ""},
				{NoteTypeFreestyle, 40, 20, 0, ""},
			}}},
		}, 30 * time.Second},
		"trailing end-of-phrase": {&Song{
			BPM: 60,
			Voices: []*Voice{{Notes: []Note{
				{NoteTypeRegular, 10, 5, 0, ""},
				{NoteTypeEndOfPhrase, 20, 0, 0, ""},
				{NoteTypeEndOfPhrase, 50, 0, 0, ""},
			}}},
		}, 15 * time.Second},
		"multiple voices with gap": {&Song{
			BPM: 120,
			Gap: 1 * time.Minute,
			Voices: []*Voice{{Notes: []Note{
				{NoteTypeGolden, 25, 30, 0, ""},
			}}, {Notes: []Note{
				{NoteTypeRap, 40, 20, 0, ""},
			}}},
		}, 90 * time.Second},
		"start and end": {&Song{
			BPM:   60,
			Gap:   1 * time.Minute,
			Start: 15 * time.Second,
			End:   2 * time.Minute,
			Voices: []*Voice{{Notes: []Note{
				{NoteTypeRegular, 0, 30, 0, ""},
				{NoteTypeRegular, 45, 30, 0, ""},
				{NoteTypeRegular, 100, 30, 0, ""},
			}}},
		}, 2*time.Minute - 15*time.Second},
		"start and irrelevant end": {&Song{
			BPM:   60,
			Gap:   1 * time.Minute,
			Start: 15 * time.Second,
			End:   2 * time.Minute,
			Voices: []*Voice{{Notes: []Note{
				{NoteTypeRegular, 0, 15, 0, ""},
			}}},
		}, 1 * time.Minute},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.song.Duration(); got != tt.want {
				t.Errorf("song.Duration() = %s, expected %s", got, tt.want)
			}
		})
	}
}

func TestSong_UpdateGap(t *testing.T) {
	song := &Song{
		BPM: 120,
		Voices: []*Voice{{Notes: []Note{
			{NoteTypeRegular, 25, 30, 0, ""},
			{NoteTypeFreestyle, 40, 20, 0, ""},
		}}},
	}
	song.UpdateGap(10 * time.Second)
	if song.Voices[P1].Notes[0].Start != 5 {
		t.Errorf("song.UpdateGap() set note start to %d, expected %d", song.Voices[P1].Notes[0].Start, 5)
	}
	if song.Voices[P1].Notes[1].Start != 20 {
		t.Errorf("song.UpdateGap() set note start to %d, expected %d", song.Voices[P1].Notes[1].Start, 40)
	}
}
