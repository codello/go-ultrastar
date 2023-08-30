package ultrastar

import (
	"testing"
	"time"
)

func TestMusic_Duration(t *testing.T) {
	ns := Notes{Note{
		Type:     NoteTypeRegular,
		Start:    120,
		Duration: 20,
		Pitch:    0,
		Text:     "text",
	}}
	expected := 1*time.Minute + 10*time.Second
	actual := ns.Duration(120)
	if actual != expected {
		t.Errorf("ns.Duration() = %s, expected %s", actual, expected)
	}
}

func TestMusic_FitBPM(t *testing.T) {
	ns := Notes{
		{NoteTypeRegular, 4, 3, 0, ""},
		{NoteTypeRegular, 8, 1, 0, ""},
		{NoteTypeRegular, 15, 4, 0, ""},
		{NoteTypeRegular, 28, 10, 0, ""},
		{NoteTypeRegular, 40, 6, 0, ""},
		{NoteTypeRegular, 56, 4, 0, ""},
	}
	oldDuration := ns.Duration(15)
	ns.ScaleBPM(15, 30)
	newDuration := ns.Duration(30)
	if oldDuration != newDuration {
		t.Errorf("ns.Duration() changed from %s to %s, expected to stay the same", oldDuration, newDuration)
	}
}
