package ultrastar

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"testing"
)

func TestNote_String(t *testing.T) {
	cases := map[string]struct {
		note     Note
		expected string
	}{
		"regular note":             {Note{NoteTypeRegular, 15, 4, 8, "go"}, ": 15 4 8 go"},
		"note with spaces in text": {Note{NoteTypeGolden, 7, 1, -2, " hey "}, "* 7 1 -2  hey "},
		"line break":               {Note{NoteTypeEndOfPhrase, 12, 7, 3, "\n"}, "- 12"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			actual := c.note.String()
			if actual != c.expected {
				t.Errorf("%v.String() = %q, expected %q", c.note, actual, c.expected)
			}
		})
	}
}

func TestNote_GobEncode(t *testing.T) {
	cases := map[string]Note{
		"regular note":             Note{NoteTypeRegular, 15, 4, 8, "go"},
		"note with spaces in text": Note{NoteTypeGolden, 7, 1, -2, " hey "},
		"note with high numbers":   Note{NoteTypeRap, 550, 20, -40, " hey "},
		"line break":               Note{NoteTypeEndOfPhrase, 12, 0, 0, "\n"},
	}
	for name, note := range cases {
		t.Run(name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			e := gob.NewEncoder(buf)
			err := e.Encode(note)
			if err != nil {
				t.Fatalf("GobEncode(%v) caused an unexpected error: %s", note, err)
			}

			var n Note
			d := gob.NewDecoder(buf)
			err = d.Decode(&n)
			if err != nil {
				t.Fatalf("GobDecode() caused an unexpected error: %s", err)
			}
			if n != note {
				t.Fatalf("GobDecode(GobEncode(%v)) = %v, expected %v", note, n, note)
			}
		})
	}
}

func ExampleNote_String() {
	n := Note{
		Type:     NoteTypeGolden,
		Start:    15,
		Duration: 4,
		Pitch:    8,
		Text:     "Go",
	}
	fmt.Println(n.String())
	// Output: * 15 4 8 Go
}
