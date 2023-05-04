package ultrastar

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMusic_Duration_Simple(t *testing.T) {
	m := NewMusicWithBPM(120)
	m.AddNote(Note{
		Type:     NoteTypeRegular,
		Start:    120,
		Duration: 20,
		Pitch:    0,
		Text:     "text",
	})
	assert.Equal(t, 1*time.Minute+10*time.Second, m.Duration())
}

func TestMusic_Duration_MultiBPM(t *testing.T) {
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
}
