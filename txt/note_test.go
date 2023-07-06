package txt

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/Karaoke-Manager/go-ultrastar"
)

func TestParseNote_success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		nType    ultrastar.NoteType
		start    ultrastar.Beat
		duration ultrastar.Beat
		pitch    ultrastar.Pitch
		text     string
	}{
		{
			name:     "regular note",
			input:    ": 5 2 3 some",
			nType:    ultrastar.NoteTypeRegular,
			start:    5,
			duration: 2,
			pitch:    3,
			text:     "some",
		},
		{
			name:     "leading space",
			input:    "* 6 2 1  body",
			nType:    ultrastar.NoteTypeGolden,
			start:    6,
			duration: 2,
			pitch:    1,
			text:     " body",
		},
		{
			name:     "trailing space",
			input:    "R 2 11 6 once  ",
			nType:    ultrastar.NoteTypeRap,
			start:    2,
			duration: 11,
			pitch:    6,
			text:     "once  ",
		},
		{
			name:     "multiple spaces",
			input:    "G  1    2    52    told ",
			nType:    ultrastar.NoteTypeGoldenRap,
			start:    1,
			duration: 2,
			pitch:    52,
			text:     "   told ",
		},
		{
			name:     "no space after note type",
			input:    "*12 41 3 me, ",
			nType:    ultrastar.NoteTypeGolden,
			start:    12,
			duration: 41,
			pitch:    3,
			text:     "me, ",
		},
	}
	for _, test := range tests {
		n, err := ParseNote(test.input)
		assert.NoError(t, err, test.name)
		assert.Equalf(t, test.start, n.Start, "%s: %s", test.name, "note start")
		assert.Equalf(t, test.duration, n.Duration, "%s: %s", test.name, "note duration")
		assert.Equalf(t, test.pitch, n.Pitch, "%s: %s", test.name, "note pitch")
		assert.Equalf(t, test.text, n.Text, "%s: %s", test.name, "note text")
	}
}

func TestParseNote_error(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "missing note type", input: ""},
		{name: "missing note start", input: ":"},
		{name: "missing note duration", input: ": 12"},
		{name: "missing note pitch", input: ": 12 3"},
		{name: "missing note text", input: ": 12 3 6"},
		{name: "float note start", input: ": 23.4 1 3 Hello"},
		{name: "invalid note type", input: "X 3 5 1 World"},
		{name: "missing space", input: ": 5 4 3test"},
	}
	for _, test := range tests {
		_, err := ParseNote(test.input)
		assert.Error(t, err, test.name)
	}
}
