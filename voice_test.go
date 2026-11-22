package ultrastar

import (
	"slices"
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
		"Empty": {Voice: Voice{}, empty: true},
		"EndOfPhrase": {Voice: Voice{Notes: []Note{
			{NoteTypeEndOfPhrase, 120, 0, 0, ""},
		}}, empty: true},
		"NotEmpty": {Voice: Voice{Notes: []Note{
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

func TestVoice_ConvertToLeadingSpaces(t *testing.T) {
	tests := map[string]struct {
		in, want []Note
	}{
		"Empty": {nil, nil},
		"MoveToNextNote": {
			[]Note{{Type: NoteTypeRegular, Text: "a  "}, {Type: NoteTypeRegular, Text: "b "}, {Type: NoteTypeRegular, Text: "c"}},
			[]Note{{Type: NoteTypeRegular, Text: "a"}, {Type: NoteTypeRegular, Text: "  b"}, {Type: NoteTypeRegular, Text: " c"}},
		},
		"LastNote": {
			[]Note{{Type: NoteTypeRegular, Text: "a "}},
			[]Note{{Type: NoteTypeRegular, Text: "a"}},
		},
		"EndOfPhrase": {
			[]Note{{Type: NoteTypeRegular, Text: "a "}, {Type: NoteTypeEndOfPhrase, Text: "\n"}, {Type: NoteTypeRegular, Text: "b"}},
			[]Note{{Type: NoteTypeRegular, Text: "a"}, {Type: NoteTypeEndOfPhrase, Text: "\n"}, {Type: NoteTypeRegular, Text: "b"}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			v := &Voice{Notes: tt.in}
			v.ConvertToLeadingSpaces()
			if !slices.Equal(v.Notes, tt.want) {
				t.Errorf("ConvertToLeadingSpaces() = %v, expected %v", v.Notes, tt.want)
			}
		})
	}
}

func TestVoice_ConvertToTrailingSpaces(t *testing.T) {
	tests := map[string]struct {
		in, want []Note
	}{
		"Empty": {nil, nil},
		"MoveToPreviousNote": {
			[]Note{{Type: NoteTypeRegular, Text: "a"}, {Type: NoteTypeRegular, Text: " b"}, {Type: NoteTypeRegular, Text: "  c"}},
			[]Note{{Type: NoteTypeRegular, Text: "a "}, {Type: NoteTypeRegular, Text: "b  "}, {Type: NoteTypeRegular, Text: "c"}},
		},
		"FirstNote": {
			[]Note{{Type: NoteTypeRegular, Text: " a"}},
			[]Note{{Type: NoteTypeRegular, Text: "a"}},
		},
		"EndOfPhrase": {
			[]Note{{Type: NoteTypeRegular, Text: "a"}, {Type: NoteTypeEndOfPhrase, Text: "\n"}, {Type: NoteTypeRegular, Text: " b"}},
			[]Note{{Type: NoteTypeRegular, Text: "a"}, {Type: NoteTypeEndOfPhrase, Text: "\n"}, {Type: NoteTypeRegular, Text: "b"}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			v := &Voice{Notes: tt.in}
			v.ConvertToTrailingSpaces()
			if !slices.Equal(v.Notes, tt.want) {
				t.Errorf("ConvertToTrailingSpaces() = %v, expected %v", v.Notes, tt.want)
			}
		})
	}
}
