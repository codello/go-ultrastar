package ultrastar

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
	"testing/iotest"
	"time"
)

func ExampleHeaderError_is() {
	err := NewHeaderError(HeaderBPM, ErrNoValue)
	fmt.Println(errors.Is(err, &HeaderError{}))
	fmt.Println(errors.Is(err, &HeaderError{Key: HeaderTitle}))
	fmt.Println(errors.Is(err, &HeaderError{Key: HeaderBPM}))
	// Output:
	// true
	// false
	// true
}

func ExampleHeaderError_as() {
	err := NewHeaderError(HeaderBPM, ErrNoValue)
	var hErr *HeaderError
	ok := errors.As(err, &hErr)
	fmt.Printf("%s: %t - %s\n", hErr.Key, ok, hErr.Err)

	hErr = &HeaderError{Key: HeaderTitle}
	ok = errors.As(err, &hErr)
	fmt.Printf("%s: %t\n", hErr.Key, ok)
	// Output:
	// BPM: true - no value
	// TITLE: false
}

func TestParseSong(t *testing.T) {
	t.Run("notes", func(t *testing.T) {
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

	t.Run("line breaks", func(t *testing.T) {
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

	t.Run("duet", func(t *testing.T) {
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

	t.Run("invalid P number", func(t *testing.T) {
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

	t.Run("stuff after end tag", func(t *testing.T) {
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

	t.Run("empty lines after tags", func(t *testing.T) {
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

	t.Run("no notes", func(t *testing.T) {
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

	t.Run("no tags", func(t *testing.T) {
		_, err := ParseSong(`: 1 2 3 some
: 4 5 6 body
* 7 8 9 once`)
		if !errors.Is(err, &HeaderError{Key: HeaderBPM}) {
			t.Errorf("ParseSong() returned an unexpected error: %s", err)
		}
	})

	t.Run("leading whitespace", func(t *testing.T) {
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

	t.Run("file without BOM", func(t *testing.T) {
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

	t.Run("file with BOM", func(t *testing.T) {
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

	t.Run("file with encoding", func(t *testing.T) {
		f, _ := os.Open("testdata/Juli - Perfekte Welle.txt")
		//goland:noinspection GoUnhandledErrorResult
		defer f.Close()
		s, err := ReadSong(f)
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

func TestSkipPrefixReader(t *testing.T) {
	tests := map[string]struct {
		input  string
		prefix string
		want   string
	}{
		"prefix present":  {"foobar", "foo", "bar"},
		"prefix absent":   {"foobar", "bar", "foobar"},
		"partial prefix":  {"foobar barfoo", "foos", "foobar barfoo"},
		"prefix complete": {"foobar", "foobar", ""},
		"longer complete": {"foo", "foobar", "foo"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			r := newSkipPrefixReader(strings.NewReader(test.input), []byte(test.prefix))
			err := iotest.TestReader(r, []byte(test.want))
			if err != nil {
				t.Errorf("skipPrefixReader(%q, %q).Read() returned an unexpected error: %s", test.input, test.prefix, err)
			}
		})
	}
}
