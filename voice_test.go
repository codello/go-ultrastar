package ultrastar

import (
	"testing"
	"time"
)

func TestVoice_LastNote(t *testing.T) {
	last := Note{NoteTypeRegular, 120, 40, 0, ""}
	v := &Voice{Notes: []Note{
		last,
		{NoteTypeEndOfPhrase, 140, 0, 0, ""},
	}}
	actual, found := v.LastNote()
	if !found {
		t.Errorf("v.LastNote() = _, false, expected true")
	}
	if actual != last {
		t.Errorf("v.LastNote() = %q, expected %q", actual, last)
	}
}

func TestVoice_Duration(t *testing.T) {
	v := &Voice{Notes: []Note{
		{NoteTypeRegular, 120, 20, 0, "text"}},
	}
	expected := 1*time.Minute + 10*time.Second
	actual := v.Duration(120)
	if actual != expected {
		t.Errorf("ns.Duration() = %s, expected %s", actual, expected)
	}
}

func TestVoice_IsEmpty(t *testing.T) {
	tests := map[string]struct {
		Voice
		empty bool
	}{
		"empty": {Voice: Voice{}, empty: true},
		"end of phrase": {Voice: Voice{Notes: []Note{
			{NoteTypeEndOfPhrase, 120, 0, 0, ""},
		}}, empty: true},
		"not empty": {Voice: Voice{Notes: []Note{
			{NoteTypeRegular, 120, 0, 0, ""},
		}}, empty: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.Voice.IsEmpty(); got != tt.empty {
				t.Errorf("v.IsEmpty() = %t, expected %t", got, tt.empty)
			}
		})
	}
}

func TestVoice_FitBPM(t *testing.T) {
	v := &Voice{Notes: []Note{
		{NoteTypeRegular, 4, 3, 0, ""},
		{NoteTypeRegular, 8, 1, 0, ""},
		{NoteTypeRegular, 15, 4, 0, ""},
		{NoteTypeRegular, 28, 10, 0, ""},
		{NoteTypeRegular, 40, 6, 0, ""},
		{NoteTypeRegular, 56, 4, 0, ""},
	}}
	oldDuration := v.Duration(15)
	v.ScaleBPM(15, 30)
	newDuration := v.Duration(30)
	if oldDuration != newDuration {
		t.Errorf("ns.Duration() changed from %s to %s, expected to stay the same", oldDuration, newDuration)
	}
}
