package txt

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"codello.dev/ultrastar"
)

func TestWriteNote(t *testing.T) {
	n := ultrastar.Note{
		Type:     ultrastar.NoteTypeRap,
		Start:    15,
		Duration: 4,
		Pitch:    -3,
		Text:     " hello ",
	}
	expected := "R 15 4 -3  hello \n"
	b := &strings.Builder{}
	err := NewWriter(b).WriteNote(n)
	actual := b.String()
	if err != nil {
		t.Errorf("WriteNote(b, %v) caused an unexpected error: %s", n, err)
	}
	if actual != expected {
		t.Errorf("WriteNote(b, %v) resulted in %q, expected %q", n, actual, expected)
	}
}

func TestWriter_WriteNote(t *testing.T) {
	n := ultrastar.Note{
		Type:     ultrastar.NoteTypeRap,
		Start:    15,
		Duration: 4,
		Pitch:    -3,
		Text:     " hello ",
	}
	expected := "R\t15\t4\t-3\t hello \n"
	b := &strings.Builder{}
	w := NewWriter(b)
	w.FieldSeparator = '\t'
	err := w.WriteNote(n)
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
		ns := ultrastar.Notes{
			{
				Type:     ultrastar.NoteTypeRegular,
				Start:    2,
				Duration: 4,
				Pitch:    8,
				Text:     "some",
			},
			{
				Type:     ultrastar.NoteTypeRegular,
				Start:    8,
				Duration: 4,
				Pitch:    8,
				Text:     "body",
			},
			{
				Type:  ultrastar.NoteTypeLineBreak,
				Start: 13,
			},
			{
				Type:     ultrastar.NoteTypeGolden,
				Start:    14,
				Duration: 4,
				Pitch:    1,
				Text:     "once",
			},
			{
				Type:     ultrastar.NoteTypeGolden,
				Start:    20,
				Duration: 4,
				Pitch:    1,
				Text:     " told",
			},
			{
				Type:     ultrastar.NoteTypeFreestyle,
				Start:    26,
				Duration: 4,
				Pitch:    1,
				Text:     " me,",
			},
		}
		b := &strings.Builder{}
		err := NewWriter(b).WriteNotes(ns)
		actual := b.String()
		expected := `: 2 4 8 some
: 8 4 8 body
- 13
* 14 4 1 once
* 20 4 1  told
F 26 4 1  me,
`
		if err != nil {
			t.Errorf("WriteNotes(b, ns) caused an unexpected error: %s", err)
		}
		if actual != expected {
			t.Errorf("WriteNotes(b, ns) resulted in %q, expected %q", actual, expected)
		}
	})
}

func TestReadWriteSong(t *testing.T) {
	f, _ := os.Open("testdata/Smash Mouth - All Star.txt")
	defer f.Close()
	expected := &bytes.Buffer{}
	s, _ := NewReader(io.TeeReader(f, expected)).ReadSong()

	actual := &strings.Builder{}
	err := WriteSong(actual, s)
	if err != nil {
		t.Errorf("WriteNotes(b, ns) caused an unexpected error: %s", err)
	}

	actualStr, expectedStr := actual.String(), expected.String()
	if actualStr != expectedStr {
		t.Errorf("WriteNotes(b, ns) resulted in %q, expected %q", actualStr, expectedStr)
	}
}
