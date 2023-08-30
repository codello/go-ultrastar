package txt

import (
	"errors"
	"os"
	"testing"
	"time"

	"codello.dev/ultrastar"
)

func TestParseSong(t *testing.T) {
	t.Run("notes", func(t *testing.T) {
		s, err := ParseSong(`#BPM:12
: 1 2 0 Some
: 3 2 0 body
`)
		if err != nil {
			t.Errorf("ParseSong() caused an unexpected error: %s", err)
		}
		if s.IsDuet() {
			t.Errorf("s.IsDuet() = true, expected false")
		}
		if s.BPM != 12*4 {
			t.Errorf("s.BPM = %f, expected %f", s.BPM, ultrastar.BPM(12*4))
		}
		if len(s.NotesP1) != 2 {
			t.Errorf("len(s.NotesP1) = %d, expected 2", len(s.NotesP1))
		}
	})

	t.Run("line breaks", func(t *testing.T) {
		s, err := ParseSong(`#BPM:4
: 1 2 4 Some
- 3
: 4 1 3 body`)
		if err != nil {
			t.Errorf("ParseSong() caused an unexpected error: %s", err)
		}
		if len(s.NotesP1) != 3 {
			t.Errorf("len(s.NotesP1) = %d, expected 3", len(s.NotesP1))
		}
	})

	t.Run("duet", func(t *testing.T) {
		s, err := ParseSong(`#BPM:2
P1
: 1 2 4 Some
P 2
: 3 4 5 body`)
		if err != nil {
			t.Errorf("ParseSong() caused an unexpected error: %s", err)
		}
		if !s.IsDuet() {
			t.Errorf("s.IsDuet() = false, expected true")
		}
		if len(s.NotesP1) != 1 {
			t.Errorf("len(s.NotesP1) = %d, expected 1", len(s.NotesP1))
		}
		if len(s.NotesP2) != 1 {
			t.Errorf("len(s.NotesP2) = %d, expected 1", len(s.NotesP2))
		}
	})

	t.Run("unexpected P number", func(t *testing.T) {
		_, err := ParseSong(`#BPM: 20
: 1 2 4 Some
P2
: 3 4 5 body`)
		if !errors.Is(err, ErrUnexpectedPNumber) {
			t.Errorf("ParseSong() did not cause ErrUnexpectedPNumber, but: %s", err)
		}
		var pErr ParseError
		errors.As(err, &pErr)
		if pErr.Line() != 3 {
			t.Errorf("pErr.Line() = %d, expected 3", pErr.Line())
		}
	})

	t.Run("invalid P number", func(t *testing.T) {
		_, err := ParseSong(`#BPM:10
P-1
: 1 2 4 Some
P2
: 3 4 5 body`)
		if !errors.Is(err, ErrInvalidPNumber) {
			t.Errorf("ParseSong() did not cause ErrInvalidPNumber, but: %s", err)
		}
		var pErr ParseError
		errors.As(err, &pErr)
		if pErr.Line() != 2 {
			t.Errorf("pErr.Line() = %d, expected 2", pErr.Line())
		}
	})

	t.Run("stuff after end tag", func(t *testing.T) {
		s, err := ParseSong(`#BPM: 42
: 1 2 4 Some
* 3 4 5 body
E
This can be anything
with multiple lines.`)
		if err != nil {
			t.Errorf("ParseSong() caused an unexpected error: %s", err)
		}
		if len(s.NotesP1) != 2 {
			t.Errorf("len(s.NotesP1) = %d, expected 2", len(s.NotesP1))
		}
	})

	t.Run("empty lines after tags", func(t *testing.T) {
		s, err := ParseSong(`#TITLE:ABC
#BPM:12

: 1 2 4 Some`)
		if err != nil {
			t.Errorf("ParseSong() caused an unexpected error: %s", err)
		}
		if len(s.NotesP1) != 1 {
			t.Errorf("len(s.NotesP1) = %d, expected 1", len(s.NotesP1))
		}
	})

	t.Run("no notes", func(t *testing.T) {
		s, err := ParseSong(`#Title:ABC
#BPM: 23`)
		if err != nil {
			t.Errorf("ParseSong() caused an unexpected error: %s", err)
		}
		if s.BPM != 23*4 {
			t.Errorf("s.BPM = %f, expected %f", s.BPM, ultrastar.BPM(23*4))
		}
		if len(s.NotesP1) != 0 {
			t.Errorf("len(s.NotesP1) = %d, expected 0", len(s.NotesP1))
		}
	})

	t.Run("no tags", func(t *testing.T) {
		s, err := ParseSong(`: 1 2 3 some
: 4 5 6 body
* 7 8 9 once`)
		if err != nil {
			t.Errorf("ParseSong() caused an unexpected error: %s", err)
		}
		if len(s.NotesP1) != 3 {
			t.Errorf("len(s.NotesP1) = %d, expected 3", len(s.NotesP1))
		}
	})

	t.Run("leading whitespace", func(t *testing.T) {
		_, err := ParseSong(`#BPM:12
: 1 2 0 Some
 : 3 2 0 body
`)
		if !errors.Is(err, ErrUnknownEvent) {
			t.Errorf("ParseSong() did not cause ErrUnknownEvent, but: %s", err)
		}
		var pErr ParseError
		errors.As(err, &pErr)
		if pErr.Line() != 3 {
			t.Errorf("pErr.Line() = %d, expected 3", pErr.Line())
		}
	})

	t.Run("file without BOM", func(t *testing.T) {
		f, _ := os.Open("testdata/Smash Mouth - All Star.txt")
		defer f.Close()
		s, err := NewReader(f).ReadSong()
		if err != nil {
			t.Errorf("ParseSong() caused an unexpected error: %s", err)
		}
		if s.Artist != "Smash Mouth" {
			t.Errorf("ParseSong() set s.Artist to %q, expected %q", s.Artist, "Smash Mouth")
		}
		if s.Year != 1999 {
			t.Errorf("ParseSong() set s.Year to %d, expected %d", s.Year, 1999)
		}
		if len(s.NotesP1) != 621 {
			t.Errorf("len(s.NotesP1) = %d, expected 621", len(s.NotesP1))
		}
		if s.NotesP1.Duration(s.BPM) != 191682692307 {
			t.Errorf("s.NotesP1.Duration() = %s, expected %s", s.NotesP1.Duration(s.BPM), time.Duration(191682692307))
		}
	})

	t.Run("file with BOM", func(t *testing.T) {
		f, _ := os.Open("testdata/Iggy Azalea - Team.txt")
		defer f.Close()
		s, err := NewReader(f).ReadSong()
		if err != nil {
			t.Errorf("ParseSong() caused an unexpected error: %s", err)
		}
		if s.Artist != "Iggy Azalea" {
			t.Errorf("ParseSong() set s.Artist to %q, expected %q", s.Artist, "Iggy Azalea")
		}
		if s.Year != 2016 {
			t.Errorf("ParseSong() set s.Year to %d, expected %d", s.Year, 2016)
		}
		if len(s.NotesP1) != 595 {
			t.Errorf("len(s.NotesP1) = %d, expected 621", len(s.NotesP1))
		}
		if s.NotesP1.Duration(s.BPM) != 196839367873 {
			t.Errorf("s.NotesP1.Duration() = %s, expected %s", s.NotesP1.Duration(s.BPM), time.Duration(196839367873))
		}
	})

	t.Run("file with encoding", func(t *testing.T) {
		f, _ := os.Open("testdata/Juli - Perfekte Welle.txt")
		defer f.Close()
		s, err := NewReader(f).ReadSong()
		if err != nil {
			t.Errorf("ParseSong() caused an unexpected error: %s", err)
		}
		if s.NotesP1[10].Text != " Träu" {
			t.Errorf("s.NotesP1[10].Text = %q, expected %q", s.NotesP1[10].Text, " Träu")
		}
		if len(s.CustomTags) != 0 {
			t.Errorf("len(s.CustomTags) = %d, expected %d", len(s.CustomTags), 0)
		}
	})
}
