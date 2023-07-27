package txt

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"codello.dev/ultrastar"
)

func TestParseSong(t *testing.T) {
	t.Run("notes", func(t *testing.T) {
		s, err := ParseSong(`#BPM:12
: 1 2 0 Some
: 3 2 0 body
`)
		assert.NoError(t, err)
		assert.False(t, s.IsDuet())
		assert.Equal(t, ultrastar.BPM(12*4), s.MusicP1.BPM())
		assert.Len(t, s.MusicP1.BPMs, 1)
		assert.Len(t, s.MusicP1.Notes, 2)
	})

	t.Run("line breaks", func(t *testing.T) {
		s, err := ParseSong(`#BPM:4
: 1 2 4 Some
- 3
: 4 1 3 body`)
		assert.NoError(t, err)
		assert.Len(t, s.MusicP1.Notes, 3)
	})

	t.Run("duet", func(t *testing.T) {
		s, err := ParseSong(`#BPM:2
P1
: 1 2 4 Some
P 2
: 3 4 5 body`)
		assert.NoError(t, err)
		assert.True(t, s.IsDuet())
		assert.Len(t, s.MusicP1.Notes, 1)
		assert.Len(t, s.MusicP2.Notes, 1)
	})

	t.Run("unexpected P number", func(t *testing.T) {
		_, err := ParseSong(`#BPM: 20
: 1 2 4 Some
P2
: 3 4 5 body`)
		var pErr ParseError
		assert.ErrorIs(t, err, ErrUnexpectedPNumber)
		assert.ErrorAs(t, err, &pErr)
		assert.Equal(t, 3, pErr.Line())
	})

	t.Run("invalid P number", func(t *testing.T) {
		_, err := ParseSong(`#BPM:10
P-1
: 1 2 4 Some
P2
: 3 4 5 body`)
		var pErr ParseError
		assert.ErrorIs(t, err, ErrInvalidPNumber)
		assert.ErrorAs(t, err, &pErr)
		assert.Equal(t, 2, pErr.Line())
	})

	t.Run("stuff after end tag", func(t *testing.T) {
		s, err := ParseSong(`#BPM: 42
: 1 2 4 Some
* 3 4 5 body
E
This can be anything
with multiple lines.`)
		assert.NoError(t, err)
		assert.Len(t, s.MusicP1.Notes, 2)
	})

	t.Run("empty lines after tags", func(t *testing.T) {
		s, err := ParseSong(`#TITLE:ABC
#BPM:12

: 1 2 4 Some`)
		assert.NoError(t, err)
		assert.Len(t, s.MusicP1.Notes, 1)
	})

	t.Run("multi BPM", func(t *testing.T) {
		s, err := ParseSong(`#BPM: 4
: 1 2 4 Some
B 5 12,3
: 10 8 1 body
B 15 1,5
: 20 1 0 once
`)
		assert.NoError(t, err)
		assert.Len(t, s.MusicP1.Notes, 3)
		assert.Len(t, s.MusicP1.BPMs, 3)
		assert.Equal(t, ultrastar.BPM(4*4), s.MusicP1.BPM())
		assert.Equal(t, ultrastar.Beat(5), s.MusicP1.BPMs[1].Start)
		assert.Equal(t, ultrastar.BPM(1.5*4), s.MusicP1.BPMs[2].BPM)
	})

	t.Run("no notes", func(t *testing.T) {
		s, err := ParseSong(`#Title:ABC
#BPM: 23`)
		assert.NoError(t, err)
		assert.Equal(t, "ABC", s.Title)
		assert.Equal(t, ultrastar.BPM(23*4), s.MusicP1.BPM())
		assert.Len(t, s.MusicP1.Notes, 0)
	})

	t.Run("no tags", func(t *testing.T) {
		s, err := ParseSong(`: 1 2 3 some
: 4 5 6 body
* 7 8 9 once`)
		assert.NoError(t, err)
		assert.Len(t, s.MusicP1.Notes, 3)
		assert.Equal(t, ultrastar.BPM(0), s.MusicP1.BPM())
	})

	t.Run("leading whitespace", func(t *testing.T) {
		_, err := ParseSong(`#BPM:12
: 1 2 0 Some
 : 3 2 0 body
`)
		var pErr ParseError
		assert.ErrorIs(t, err, ErrUnknownEvent)
		assert.ErrorAs(t, err, &pErr)
		assert.Equal(t, 3, pErr.Line())
	})

	t.Run("file without BOM", func(t *testing.T) {
		f, _ := os.Open("testdata/Smash Mouth - All Star.txt")
		defer f.Close()
		s, err := ReadSong(f)
		require.NoError(t, err)
		assert.False(t, s.IsDuet())
		assert.Equal(t, "Smash Mouth", s.Artist)
		assert.Equal(t, 1999, s.Year)
		assert.Len(t, s.MusicP1.Notes, 621)
		assert.Equal(t, ultrastar.BPM(312*4), s.BPM())
		assert.Equal(t, time.Duration(191682692307), s.MusicP1.Duration())
	})

	t.Run("file with BOM", func(t *testing.T) {
		f, _ := os.Open("testdata/Iggy Azalea - Team.txt")
		defer f.Close()
		s, err := ReadSong(f)
		require.NoError(t, err)
		assert.False(t, s.IsDuet())
		assert.Equal(t, "Iggy Azalea", s.Artist)
		assert.Equal(t, 2016, s.Year)
		assert.Len(t, s.MusicP1.Notes, 595)
		assert.Equal(t, ultrastar.BPM(199.96*4), s.BPM())
		assert.Equal(t, time.Duration(196839367873), s.MusicP1.Duration())
	})

	t.Run("file with encoding", func(t *testing.T) {
		f, _ := os.Open("testdata/Juli - Perfekte Welle.txt")
		defer f.Close()
		s, err := ReadSong(f)
		require.NoError(t, err)

		assert.Equal(t, " Träu", s.MusicP1.Notes[10].Text)
		assert.Empty(t, s.CustomTags)
	})
}
