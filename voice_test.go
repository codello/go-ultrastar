package ultrastar

import (
	"testing"
	"time"
)

func TestMusic_Duration(t *testing.T) {
	v := &Voice{Notes: []Note{
		{NoteTypeRegular, 120, 20, 0, "text"}},
	}
	expected := 1*time.Minute + 10*time.Second
	actual := v.Duration(120)
	if actual != expected {
		t.Errorf("ns.Duration() = %s, expected %s", actual, expected)
	}
}

func TestMusic_FitBPM(t *testing.T) {
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
