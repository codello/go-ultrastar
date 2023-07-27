package txt

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Karaoke-Manager/go-ultrastar"
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
	err := FormatDefault.WriteNote(b, n)
	assert.NoError(t, err)
	assert.Equal(t, expected, b.String())
}

func TestFormat_WriteNote(t *testing.T) {
	n := ultrastar.Note{
		Type:     ultrastar.NoteTypeRap,
		Start:    15,
		Duration: 4,
		Pitch:    -3,
		Text:     " hello ",
	}
	expected := "R\t15\t4\t-3\t hello \n"
	b := &strings.Builder{}
	f := &Format{
		FieldSeparator: '\t',
	}
	err := f.WriteNote(b, n)
	assert.NoError(t, err)
	assert.Equal(t, expected, b.String())
}

func TestWriteMusic(t *testing.T) {
	t.Run("music formatting", func(t *testing.T) {
		m := &ultrastar.Music{
			Notes: []ultrastar.Note{
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
			},
			BPMs: []ultrastar.BPMChange{{
				Start: 0,
				BPM:   120,
			}, {
				Start: 22,
				BPM:   50,
			}},
		}
		b := &strings.Builder{}
		err := FormatDefault.WriteMusic(b, m)
		assert.NoError(t, err)
		assert.Equal(t, `B 0 120
: 2 4 8 some
: 8 4 8 body
- 13
* 14 4 1 once
* 20 4 1  told
B 22 50
F 26 4 1  me,
`, b.String())
	})
}

func TestReadWriteSong(t *testing.T) {
	f, _ := os.Open("testdata/Smash Mouth - All Star.txt")
	defer f.Close()
	expected := &bytes.Buffer{}
	s, _ := ReadSong(io.TeeReader(f, expected))

	actual := &strings.Builder{}
	err := WriteSong(actual, s)
	assert.NoError(t, err)
	assert.Equal(t, expected.String(), actual.String())
}
