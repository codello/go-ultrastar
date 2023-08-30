package txt

import (
	"testing"

	"codello.dev/ultrastar"
)

func TestParseNote(t *testing.T) {
	cases := map[string]struct {
		input    string
		expected ultrastar.Note
		error    bool
	}{
		"regular note":             {": 5 2 3 some", ultrastar.Note{Type: ultrastar.NoteTypeRegular, Start: 5, Duration: 2, Pitch: 3, Text: "some"}, false},
		"leading space":            {"* 6 2 1  body", ultrastar.Note{Type: ultrastar.NoteTypeGolden, Start: 6, Duration: 2, Pitch: 1, Text: " body"}, false},
		"trailing space":           {"R 2 11 6 once  ", ultrastar.Note{Type: ultrastar.NoteTypeRap, Start: 2, Duration: 11, Pitch: 6, Text: "once  "}, false},
		"multiple spaces":          {"G  1    2    52    told ", ultrastar.Note{Type: ultrastar.NoteTypeGoldenRap, Start: 1, Duration: 2, Pitch: 52, Text: "   told "}, false},
		"no space after note type": {"*12 41 3 me, ", ultrastar.Note{Type: ultrastar.NoteTypeGolden, Start: 12, Duration: 41, Pitch: 3, Text: "me, "}, false},
		"missing note type":        {"", ultrastar.Note{}, true},
		"missing note start":       {":", ultrastar.Note{Type: ultrastar.NoteTypeRegular}, true},
		"missing note duration":    {": 12", ultrastar.Note{Type: ultrastar.NoteTypeRegular, Start: 12}, true},
		"missing note pitch":       {": 12 3", ultrastar.Note{Type: ultrastar.NoteTypeRegular, Start: 12, Duration: 3}, true},
		"missing note text":        {": 12 3 6", ultrastar.Note{Type: ultrastar.NoteTypeRegular, Start: 12, Duration: 3, Pitch: 6}, true},
		"float note start":         {": 23.4 1 3 Hello", ultrastar.Note{Type: ultrastar.NoteTypeRegular}, true},
		"invalid note type":        {"X 3 5 1 World", ultrastar.Note{}, true},
		"missing space":            {": 5 4 3test", ultrastar.Note{Type: ultrastar.NoteTypeRegular, Start: 5, Duration: 4}, true},
		"line break":               {"- 52", ultrastar.Note{Type: ultrastar.NoteTypeLineBreak, Start: 52, Text: "\n"}, false},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			actual, err := ParseNote(c.input)
			if err != nil && !c.error {
				t.Errorf("ParseNote(%q) returned an unexpected error: %s", c.input, err)
			} else if err == nil && c.error {
				t.Errorf("ParseNote(%q) did not return an error, but one was expected", c.input)
			}
			if actual != c.expected {
				t.Errorf("ParseNote(%q) = %v, expected %v", c.input, actual, c.expected)
			}
		})
	}
}
