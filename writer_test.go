package ultrastar

import (
	"bytes"
	"io"
	"os"
	"slices"
	"strings"
	"testing"
	"time"
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

func TestWriteSong(t *testing.T) {
	t.Run("headers", func(t *testing.T) {
		s := &Song{
			BPM:    100,
			Artist: []string{"A, B"},
			Voices: []*Voice{{Name: "Voice 1"}, {Name: "Voice 2"}},
		}
		var b strings.Builder
		if err := WriteSong(&b, s, Version200); err != nil {
			t.Fatalf("WriteSong() returned an unexpected error: %s", err)
		}
		out := b.String()
		for _, line := range []string{"#ARTIST:A,, B\n", "#P1:Voice 1\n", "#P2:Voice 2\n"} {
			if !strings.Contains(out, line) {
				t.Errorf("WriteSong() output does not contain %q:\n%s", line, out)
			}
		}
		if strings.Contains(out, "#YEAR") {
			t.Errorf("WriteSong() wrote a #YEAR header for a song without a year:\n%s", out)
		}
		if !slices.Equal(s.Artist, []string{"A, B"}) {
			t.Errorf("WriteSong() changed s.Artist to %q", s.Artist)
		}
	})

	t.Run("custom medley headers", func(t *testing.T) {
		s := &Song{BPM: 100, Header: Header{HeaderMedleyStartBeat: {"10"}}}
		var b strings.Builder
		if err := WriteSong(&b, s, Version200); err != nil {
			t.Fatalf("WriteSong() returned an unexpected error: %s", err)
		}
		if strings.Contains(b.String(), HeaderMedleyStartBeat) {
			t.Errorf("WriteSong() wrote the standard header %s from s.Header:\n%s", HeaderMedleyStartBeat, b.String())
		}
	})

	t.Run("medley beats", func(t *testing.T) {
		s := &Song{BPM: 400, Gap: 2 * time.Second, MedleyStart: 5 * time.Second, MedleyEnd: 8 * time.Second}
		var b strings.Builder
		if err := WriteSong(&b, s, Version120); err != nil {
			t.Fatalf("WriteSong() returned an unexpected error: %s", err)
		}
		s2, err := ParseSong(b.String())
		if err != nil {
			t.Fatalf("ParseSong() returned an unexpected error: %s", err)
		}
		if s2.MedleyStart != s.MedleyStart || s2.MedleyEnd != s.MedleyEnd {
			t.Errorf("medley = %s-%s, expected %s-%s", s2.MedleyStart, s2.MedleyEnd, s.MedleyStart, s.MedleyEnd)
		}
	})
}

func TestWriter_zeroVersion(t *testing.T) {
	s := &Song{BPM: 100, Voices: []*Voice{{Name: "1"}, {Name: "2"}, {Name: "3"}}}
	var b strings.Builder
	w := NewWriter(&b, Version{})
	w.Version = Version{}
	if err := w.WriteSong(s); err != nil {
		t.Fatalf("WriteSong() returned an unexpected error: %s", err)
	}
	if !strings.HasPrefix(b.String(), "#VERSION:2.0.0\n") {
		t.Errorf("WriteSong() did not write version 2.0.0:\n%s", b.String())
	}
}
