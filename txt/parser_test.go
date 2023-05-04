package txt

import (
	"github.com/codello/ultrastar"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseSong_Notes(t *testing.T) {
	s, err := ParseSong(`#BPM:12
: 1 2 0 Some
: 3 2 0 body
`)
	assert.NoError(t, err)
	assert.False(t, s.IsDuet())
	assert.Equal(t, ultrastar.BPM(12*4), s.MusicP1.BPM())
	assert.Len(t, s.MusicP1.BPMs, 1)
	assert.Len(t, s.MusicP1.Notes, 2)
	assert.Len(t, s.MusicP1.LineBreaks, 0)
}

func TestParseSong_LineBreaks(t *testing.T) {
	s, err := ParseSong(`#BPM:4
: 1 2 4 Some
- 3
: 4 1 3 body`)
	assert.NoError(t, err)
	assert.Len(t, s.MusicP1.Notes, 2)
	assert.Len(t, s.MusicP1.LineBreaks, 1)
}

func TestParseSong_Duet(t *testing.T) {
	s, err := ParseSong(`#BPM:2
P1
: 1 2 4 Some
P 2
: 3 4 5 body`)
	assert.NoError(t, err)
	assert.True(t, s.IsDuet())
	assert.Len(t, s.MusicP1.Notes, 1)
	assert.Len(t, s.MusicP2.Notes, 1)
}

func TestParseSong_UnexpectedPNumber(t *testing.T) {
	_, err := ParseSong(`#BPM: 20
: 1 2 4 Some
P2
: 3 4 5 body`)
	var pErr *ParseError
	assert.ErrorIs(t, err, ErrUnexpectedPNumber)
	assert.ErrorAs(t, err, &pErr)
	assert.Equal(t, 3, pErr.Line())
}

func TestParseSong_InvalidPNumber(t *testing.T) {
	_, err := ParseSong(`#BPM:10
P-1
: 1 2 4 Some
P2
: 3 4 5 body`)
	var pErr *ParseError
	assert.ErrorIs(t, err, ErrInvalidPNumber)
	assert.ErrorAs(t, err, &pErr)
	assert.Equal(t, 2, pErr.Line())
}

func TestParseSong_StuffAfterEndTag(t *testing.T) {
	s, err := ParseSong(`#BPM: 42
: 1 2 4 Some
* 3 4 5 body
E
This can be anything
with multiple lines.`)
	assert.NoError(t, err)
	assert.Len(t, s.MusicP1.Notes, 2)
}

func TestParseSong_EmptyLinesAfterTags(t *testing.T) {
	s, err := ParseSong(`#TITLE:ABC
#BPM:12

: 1 2 4 Some`)
	assert.NoError(t, err)
	assert.Len(t, s.MusicP1.Notes, 1)
}

func TestParseSong_MultiBPM(t *testing.T) {
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
}

func TestParseSong_NoNotes(t *testing.T) {
	s, err := ParseSong(`#Title:ABC
#BPM: 23`)
	assert.NoError(t, err)
	assert.Equal(t, "ABC", s.Title)
	assert.Equal(t, ultrastar.BPM(23*4), s.MusicP1.BPM())
	assert.Len(t, s.MusicP1.Notes, 0)
}

func TestParseSong_NoTags(t *testing.T) {
	s, err := ParseSong(`: 1 2 3 some
: 4 5 6 body
* 7 8 9 once`)
	assert.NoError(t, err)
	assert.Len(t, s.MusicP1.Notes, 3)
	assert.Equal(t, ultrastar.BPM(0), s.MusicP1.BPM())
}

func TestParseSong_LeadingWhitespace(t *testing.T) {
	_, err := ParseSong(`#BPM:12
: 1 2 0 Some
 : 3 2 0 body
`)
	var pErr *ParseError
	assert.ErrorIs(t, err, ErrUnknownEvent)
	assert.ErrorAs(t, err, &pErr)
	assert.Equal(t, 3, pErr.Line())
}

// TODO: Probably more tests
