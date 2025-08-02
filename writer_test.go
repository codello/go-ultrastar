package ultrastar

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestWriteNote(t *testing.T) {
	n := Note{
		Type:     NoteTypeRap,
		Start:    15,
		Duration: 4,
		Pitch:    -3,
		Text:     " hello ",
	}
	expected := "R 15 4 -3  hello \n"
	b := &strings.Builder{}
	err := NewWriter(b, Version120).WriteNote(n, P1)
	actual := b.String()
	if err != nil {
		t.Errorf("WriteNote(b, %v) caused an unexpected error: %s", n, err)
	}
	if actual != expected {
		t.Errorf("WriteNote(b, %v) resulted in %q, expected %q", n, actual, expected)
	}
}

func TestWriter_WriteNote(t *testing.T) {
	n := Note{
		Type:     NoteTypeRap,
		Start:    15,
		Duration: 4,
		Pitch:    -3,
		Text:     " hello ",
	}
	expected := "R 15 4 -3  hello \n"
	b := &strings.Builder{}
	w := NewWriter(b, Version120)
	err := w.WriteNote(n, P1)
	actual := b.String()
	if err != nil {
		t.Errorf("WriteNote(b, %v) caused an unexpected error: %s", n, err)
	}
	if actual != expected {
		t.Errorf("WriteNote(b, %v) resulted in %q, expected %q", n, actual, expected)
	}
}

func TestWriteNotes(t *testing.T) {
	t.Run("notes formatting", func(t *testing.T) {
		ns := Voice{
			Name: "A Voice",
			Notes: []Note{
				{
					Type:     NoteTypeRegular,
					Start:    2,
					Duration: 4,
					Pitch:    8,
					Text:     "some",
				},
				{
					Type:     NoteTypeRegular,
					Start:    8,
					Duration: 4,
					Pitch:    8,
					Text:     "body",
				},
				{
					Type:  NoteTypeEndOfPhrase,
					Start: 13,
				},
				{
					Type:     NoteTypeGolden,
					Start:    14,
					Duration: 4,
					Pitch:    1,
					Text:     "once",
				},
				{
					Type:     NoteTypeGolden,
					Start:    20,
					Duration: 4,
					Pitch:    1,
					Text:     " told",
				},
				{
					Type:     NoteTypeFreestyle,
					Start:    26,
					Duration: 4,
					Pitch:    1,
					Text:     " me,",
				},
			},
		}
		b := &strings.Builder{}
		w := NewWriter(b, Version120)
		for _, n := range ns.Notes {
			if err := w.WriteNote(n, P1); err != nil {
				t.Errorf("WriteNote(n, 1) caused an unexpected error: %s", err)
				return
			}
		}
		actual := b.String()
		expected := `: 2 4 8 some
: 8 4 8 body
- 13
* 14 4 1 once
* 20 4 1  told
F 26 4 1  me,
`
		if actual != expected {
			t.Errorf("WriteNotes(b, ns) resulted in %q, expected %q", actual, expected)
		}
	})
}

func TestReadWriteSong(t *testing.T) {
	f, _ := os.Open("testdata/Smash Mouth - All Star.txt")
	defer f.Close()
	expected := &bytes.Buffer{}
	r, _ := NewReader(io.TeeReader(f, expected))
	s, _ := r.ReadSong()

	actual := &strings.Builder{}
	err := WriteSong(actual, s, r.Version)
	if err != nil {
		t.Errorf("WriteSong(b, ns) caused an unexpected error: %s", err)
	}

	actualStr, expectedStr := actual.String(), expected.String()
	if actualStr != expectedStr {
		t.Errorf("WriteNotes(b, ns) resulted in %q, expected %q", actualStr, expectedStr)
	}
}
