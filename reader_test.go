package ultrastar

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"testing"
	"time"
)

func ExampleHeaderError() {
	_, err := ParseSong(": 1 2 3 some\n")
	var hErr *HeaderError
	if errors.As(err, &hErr) {
		fmt.Printf("%s: %s\n", hErr.Key, hErr.Err)
	}
	// Output:
	// BPM: no value
}

func TestParseSong(t *testing.T) {
	t.Run("Notes", func(t *testing.T) {
		s, err := ParseSong(`#BPM:12
: 1 2 0 Some
: 3 2 0 body
`)
		if err != nil {
			t.Errorf("ParseSong() returned an unexpected error: %s", err)
			return
		}
		if s.IsDuet() {
			t.Errorf("s.IsDuet() = true, expected false")
		}
		if s.BPM != 12*4 {
			t.Errorf("s.BPM = %f, expected %f", s.BPM, BPM(12*4))
		}
		if len(s.Voices[P1].Notes) != 2 {
			t.Errorf("len(s.Voices[P1].Notes) = %d, expected 2", len(s.Voices[P1].Notes))
		}
	})

	t.Run("LineBreaks", func(t *testing.T) {
		s, err := ParseSong(`#BPM:4
: 1 2 4 Some
- 3
: 4 1 3 body`)
		if err != nil {
			t.Errorf("ParseSong() returned an unexpected error: %s", err)
		}
		if len(s.Voices[P1].Notes) != 3 {
			t.Errorf("len(s.Voices[P1].Notes) = %d, expected 3", len(s.Voices[P1].Notes))
		}
	})

	t.Run("Duet", func(t *testing.T) {
		s, err := ParseSong(`#BPM:2
P1
: 1 2 4 Some
P 2
: 3 4 5 body`)
		if err != nil {
			t.Errorf("ParseSong() returned an unexpected error: %s", err)
		}
		if !s.IsDuet() {
			t.Errorf("s.IsDuet() = false, expected true")
		}
		if len(s.Voices[P1].Notes) != 1 {
			t.Errorf("len(s.Voices[P1].Notes) = %d, expected 1", len(s.Voices[P1].Notes))
		}
		if len(s.Voices[P2].Notes) != 1 {
			t.Errorf("len(s.Voices[P2].Notes) = %d, expected 1", len(s.Voices[P2].Notes))
		}
	})

	t.Run("InvalidVoiceChange", func(t *testing.T) {
		_, err := ParseSong(`#BPM:10
P-1
: 1 2 4 Some
P2
: 3 4 5 body`)
		var sErr *SyntaxError
		if !errors.As(err, &sErr) {
			t.Errorf("ParseSong() returned an unexpected error: %s", err)
		}
		if sErr.Line != 2 {
			t.Errorf("sErr.Line = %d, expected 2", sErr.Line)
		}
	})

	t.Run("TextAfterEndTag", func(t *testing.T) {
		s, err := ParseSong(`#BPM: 42
: 1 2 4 Some
* 3 4 5 body
E
This can be anything
with multiple lines.`)
		if err != nil {
			t.Errorf("ParseSong() returned an unexpected error: %s", err)
		}
		if len(s.Voices[P1].Notes) != 2 {
			t.Errorf("len(s.Voices[P1].Notes) = %d, expected 2", len(s.Voices[P1].Notes))
		}
	})

	t.Run("EmptyLinesAfterTags", func(t *testing.T) {
		s, err := ParseSong(`#TITLE:ABC
#BPM:12

: 1 2 4 Some`)
		if err != nil {
			t.Errorf("ParseSong() returned an unexpected error: %s", err)
		}
		if len(s.Voices[P1].Notes) != 1 {
			t.Errorf("len(s.Voices[P1].Notes) = %d, expected 1", len(s.Voices[P1].Notes))
		}
	})

	t.Run("NoNotes", func(t *testing.T) {
		s, err := ParseSong(`#Title:ABC
#BPM: 23`)
		if err != nil {
			t.Errorf("ParseSong() returned an unexpected error: %s", err)
			return
		}
		if s.BPM != 23*4 {
			t.Errorf("s.BPM = %f, expected %f", s.BPM, BPM(23*4))
		}
		if len(s.Voices) != 0 {
			t.Errorf("len(s.Voices) = %d, expected 0", len(s.Voices))
		}
	})

	t.Run("NoTags", func(t *testing.T) {
		_, err := ParseSong(`: 1 2 3 some
: 4 5 6 body
* 7 8 9 once`)
		var hErr *HeaderError
		if !errors.As(err, &hErr) || hErr.Key != HeaderBPM {
			t.Errorf("ParseSong() returned an unexpected error: %s", err)
		}
	})

	t.Run("LeadingWhitespace", func(t *testing.T) {
		_, err := ParseSong(`#BPM:12
: 1 2 0 Some
 : 3 2 0 body
`)
		var sErr *SyntaxError
		if !errors.As(err, &sErr) {
			t.Errorf("ParseSong() returned an unexpected error: %s", err)
		}
		if sErr.Line != 3 {
			t.Errorf("pErr.Line() = %d, expected 3", sErr.Line)
		}
	})

	t.Run("FileWithoutBOM", func(t *testing.T) {
		f, _ := os.Open("testdata/Smash Mouth - All Star.txt")
		//goland:noinspection GoUnhandledErrorResult
		defer f.Close()
		s, err := ReadSong(f)
		if err != nil {
			t.Errorf("ReadSong() returned an unexpected error: %s", err)
			return
		}
		if !slices.Equal(s.Artist, []string{"Smash Mouth"}) {
			t.Errorf("ReadSong() set Artist to %q, expected %q", s.Artist, "Smash Mouth")
		}
		if s.Year != 1999 {
			t.Errorf("ReadSong() set Year to %d, expected %d", s.Year, 1999)
		}
		if len(s.Voices[P1].Notes) != 621 {
			t.Errorf("len(s.Voices[P1].Notes) = %d, expected 621", len(s.Voices[P1].Notes))
		}
		if s.Voices[P1].Duration(s.BPM) != 191682692307 {
			t.Errorf("s.Voices[P1].Duration() = %s, expected %s", s.Voices[P1].Duration(s.BPM), time.Duration(191682692307))
		}
	})

	t.Run("FileWithBOM", func(t *testing.T) {
		f, _ := os.Open("testdata/Iggy Azalea - Team.txt")
		//goland:noinspection GoUnhandledErrorResult
		defer f.Close()
		s, err := ReadSong(f)
		if err != nil {
			t.Errorf("ReadSong() returned an unexpected error: %s", err)
			return
		}
		if !slices.Equal(s.Artist, []string{"Iggy Azalea"}) {
			t.Errorf("ReadSong() set Artist to %q, expected %q", s.Artist, "Iggy Azalea")
		}
		if s.Year != 2016 {
			t.Errorf("ReadSong() set Year to %d, expected %d", s.Year, 2016)
		}
		if len(s.Voices[P1].Notes) != 595 {
			t.Errorf("len(s.NotesP1) = %d, expected 621", len(s.Voices[P1].Notes))
		}
		if s.Voices[P1].Duration(s.BPM) != 196839367873 {
			t.Errorf("s.Voices[P1].Duration() = %s, expected %s", s.Voices[P1].Duration(s.BPM), time.Duration(196839367873))
		}
	})

	t.Run("FileWithEncoding", func(t *testing.T) {
		f, _ := os.Open("testdata/Juli - Perfekte Welle.txt")
		//goland:noinspection GoUnhandledErrorResult
		defer f.Close()
		r := &Reader{EncodingReader: latin1Reader}
		err := r.Reset(f)
		if err != nil {
			t.Errorf("Reset() returned an unexpected error: %s", err)
			return
		}
		s, err := r.ReadSong()
		if err != nil {
			t.Errorf("ParseSong() returned an unexpected error: %s", err)
			return
		}
		if s.Voices[P1].Notes[10].Text != "Träu" {
			t.Errorf("s.Voices[P1].Notes[10].Text = %q, expected %q", s.Voices[P1].Notes[10].Text, "Träu")
		}
		if len(s.Header) != 1 {
			t.Errorf("len(s.Header) = %d, expected %d", len(s.Header), 1)
		}
	})
}

func TestReader_Song(t *testing.T) {
	t.Run("VoiceNames", func(t *testing.T) {
		r, err := NewReader(strings.NewReader("#BPM:12\n#P1:A\n#P2:B\n"))
		if err != nil {
			t.Fatalf("NewReader() returned an unexpected error: %s", err)
		}
		s, err := r.Song()
		if err != nil {
			t.Fatalf("Song() returned an unexpected error: %s", err)
		}
		if len(s.Voices) != 2 {
			t.Fatalf("len(s.Voices) = %d, expected 2", len(s.Voices))
		}
		if s.Voices[P1].Name != "A" || s.Voices[P2].Name != "B" {
			t.Errorf("voice names = %q, %q, expected %q, %q", s.Voices[P1].Name, s.Voices[P2].Name, "A", "B")
		}
	})

	t.Run("BPM", func(t *testing.T) {
		tests := map[string]BPM{
			"#BPM:100\n":                 400,
			"#VERSION:1.0.0\n#BPM:100\n": 400,
			"#VERSION:2.0.0\n#BPM:100\n": 100,
		}
		for in, expected := range tests {
			s, err := ParseSong(in)
			if err != nil {
				t.Errorf("ParseSong(%q) returned an unexpected error: %s", in, err)
				continue
			}
			if s.BPM != expected {
				t.Errorf("ParseSong(%q).BPM = %f, expected %f", in, s.BPM, expected)
			}
		}
	})
}

// latin1Reader is an EncodingReader that supports ISO 8859-1 under the name
// "CP1252". This is sufficient for the test data.
func latin1Reader(encoding string, input io.Reader) (io.Reader, error) {
	if !strings.EqualFold(encoding, "cp1252") {
		return nil, fmt.Errorf("unsupported encoding %q", encoding)
	}
	b, err := io.ReadAll(input)
	if err != nil {
		return nil, err
	}
	runes := make([]rune, len(b))
	for i, c := range b {
		runes[i] = rune(c)
	}
	return strings.NewReader(string(runes)), nil
}

func TestReader_encoding(t *testing.T) {
	const song = "#ENCODING:CP1252\n#TITLE:Tr\xe4ume\n#BPM:12\n: 1 2 3 Tr\xe4u\n"

	t.Run("NoEncodingReader", func(t *testing.T) {
		r, err := NewReader(strings.NewReader(song))
		var hErr *HeaderError
		if !errors.As(err, &hErr) || hErr.Key != HeaderEncoding {
			t.Fatalf("NewReader() returned %v, expected a HeaderError for %s", err, HeaderEncoding)
		}
		if r.Encoding != "" {
			t.Errorf("r.Encoding = %q, expected %q", r.Encoding, "")
		}
		n, _, err := r.ReadNote()
		if err != nil {
			t.Fatalf("ReadNote() returned an unexpected error: %s", err)
		}
		if n.Text != "Tr\xe4u" {
			t.Errorf("n.Text = %q, expected the raw text %q", n.Text, "Tr\xe4u")
		}
	})

	t.Run("UseEncoding", func(t *testing.T) {
		r := &Reader{EncodingReader: latin1Reader}
		if err := r.Reset(strings.NewReader("#TITLE:Tr\xe4ume\n#BPM:12\n: 1 2 3 Tr\xe4u\n")); err != nil {
			t.Fatalf("Reset() returned an unexpected error: %s", err)
		}
		if err := r.UseEncoding("CP1252"); err != nil {
			t.Fatalf("UseEncoding() returned an unexpected error: %s", err)
		}
		if title := r.Header.Get(HeaderTitle); title != "Träume" {
			t.Errorf("r.Header.Get(%q) = %q, expected %q", HeaderTitle, title, "Träume")
		}
		n, _, err := r.ReadNote()
		if err != nil {
			t.Fatalf("ReadNote() returned an unexpected error: %s", err)
		}
		if n.Text != "Träu" {
			t.Errorf("n.Text = %q, expected %q", n.Text, "Träu")
		}
		if err := r.UseEncoding("unknown"); err == nil {
			t.Errorf("UseEncoding(%q) returned no error", "unknown")
		}
		if r.Encoding != "CP1252" {
			t.Errorf("r.Encoding = %q after a failed UseEncoding, expected %q", r.Encoding, "CP1252")
		}
	})
}

func TestParseSong_unsupportedEncoding(t *testing.T) {
	s, err := ParseSong("#ENCODING:CP1252\n#BPM:12\n: 1 2 3 Tr\xe4u\n")
	var hErr *HeaderError
	if !errors.As(err, &hErr) || hErr.Key != HeaderEncoding {
		t.Fatalf("ParseSong() returned %v, expected a HeaderError for %s", err, HeaderEncoding)
	}
	if len(s.Voices) != 1 || len(s.Voices[P1].Notes) != 1 {
		t.Fatalf("ParseSong() did not read the notes: %+v", s.Voices)
	}
	if text := s.Voices[P1].Notes[0].Text; text != "Tr\xe4u" {
		t.Errorf("note text = %q, expected the raw text %q", text, "Tr\xe4u")
	}
}
