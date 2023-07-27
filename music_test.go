package ultrastar

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMusic_Duration(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		m := NewMusicWithBPM(120)
		m.AddNote(Note{
			Type:     NoteTypeRegular,
			Start:    120,
			Duration: 20,
			Pitch:    0,
			Text:     "text",
		})
		assert.Equal(t, 1*time.Minute+10*time.Second, m.Duration())
	})

	t.Run("multi BPM", func(t *testing.T) {
		m := NewMusicWithBPM(120)
		m.BPMs = append(m.BPMs, BPMChange{
			Start: 60,
			BPM:   30,
		})
		m.AddNote(Note{
			Type:     NoteTypeRegular,
			Start:    100,
			Duration: 20,
			Pitch:    0,
			Text:     "text",
		})
		// 60 beats at 120 BPM, 60 beats at 30 BPM
		assert.Equal(t, 2*time.Minute+30*time.Second, m.Duration())
	})
}

func TestMusic_FitBPM(t *testing.T) {
	m := NewMusic()
	m.BPMs = []BPMChange{
		{0, 15},
		{20, 60},
		{50, 20},
	}
	m.Notes = Notes{
		{NoteTypeRegular, 4, 3, 0, ""},
		{NoteTypeRegular, 8, 1, 0, ""},
		{NoteTypeRegular, 15, 4, 0, ""},
		{NoteTypeRegular, 28, 10, 0, ""},
		{NoteTypeRegular, 40, 6, 0, ""},
		{NoteTypeRegular, 56, 8, 0, ""},
	}
	oldDuration := m.Duration()
	m.FitBPM(30)
	assert.Equal(t, oldDuration, m.Duration(), "m.Duration()")
	assert.Equal(t, Notes{
		{NoteTypeRegular, 8, 6, 0, ""},
		{NoteTypeRegular, 16, 2, 0, ""},
		{NoteTypeRegular, 30, 8, 0, ""},
		{NoteTypeRegular, 44, 5, 0, ""},
		{NoteTypeRegular, 50, 3, 0, ""},
		{NoteTypeRegular, 64, 12, 0, ""},
	}, m.Notes, "m.Notes")
	assert.Equal(t, []BPMChange{{0, 30}}, m.BPMs, "m.BPMs")
}

func TestMusic_BPM(t *testing.T) {
	cases := map[string]struct {
		changes  []BPMChange
		expected BPM
		invalid  bool
	}{
		"single":          {[]BPMChange{{0, 15}}, 15, false},
		"multi":           {[]BPMChange{{0, 80}, {20, 30}}, 80, false},
		"no starting BPM": {[]BPMChange{{10, 30}}, 0, true},
		"empty BPMs":      {nil, 0, true},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			m := NewMusic()
			m.BPMs = c.changes
			bpm := m.BPM()
			if c.invalid {
				assert.False(t, bpm.IsValid())
			} else {
				assert.True(t, bpm.IsValid())
				assert.Equal(t, c.expected, bpm)
			}
		})
	}
}
