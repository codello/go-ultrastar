package txt

import (
	"github.com/codello/ultrastar"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// TestSetTag_StringMetadata tests that known string tags are stored correctly.
func TestSetTag_StringMetadata(t *testing.T) {
	s := ultrastar.NewSong()
	tests := map[string]*string{
		TagMP3:        &s.AudioFile,
		TagVideo:      &s.VideoFile,
		TagCover:      &s.CoverFile,
		TagBackground: &s.BackgroundFile,

		TagTitle:    &s.Title,
		TagArtist:   &s.Artist,
		TagGenre:    &s.Genre,
		TagEdition:  &s.Edition,
		TagCreator:  &s.Creator,
		TagAuthor:   &s.Creator,
		TagLanguage: &s.Language,

		TagComment:      &s.Comment,
		TagDuetSingerP1: &s.DuetSinger1,
		TagDuetSingerP2: &s.DuetSinger2,
		TagP1:           &s.DuetSinger1,
		TagP2:           &s.DuetSinger2,
	}
	for tag, field := range tests {
		err := SetTag(s, tag, "Hello World")
		assert.NoError(t, err, tag)
		assert.Equal(t, "Hello World", *field, tag)
		assert.Empty(t, s.CustomTags, tag)
		// Reset field for next test
		*field = ""
	}
}

// TestSetTag_CustomTag tests that custom tags are stored as their upper case
// version.
func TestSetTag_CustomTag(t *testing.T) {
	s := ultrastar.NewSong()
	err := SetTag(s, "My Tag", "My Value")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(s.CustomTags))
	assert.NotContains(t, s.CustomTags, "My Tag")
	assert.Contains(t, s.CustomTags, "MY TAG")
	assert.Equal(t, "My Value", s.CustomTags["MY TAG"])
}

// TestSetTag_UpperCase tests that known tags are processed case insensitively.
func TestSetTag_UpperCase(t *testing.T) {
	s := ultrastar.NewSong()
	err := SetTag(s, "title", "Some Value")
	assert.NoError(t, err)
	assert.Equal(t, "Some Value", s.Title)
}

func TestSetTag_Gap(t *testing.T) {
	s := ultrastar.NewSong()
	err := SetTag(s, TagGap, "123.24")
	assert.NoError(t, err)
	expected := time.Duration(123.24 * float64(time.Millisecond))
	assert.Equal(t, expected, s.Gap)
}

func TestSetTag_InvalidGap(t *testing.T) {
	tests := []struct {
		name string
		test string
	}{
		{test: "31abc", name: "letters"},
		{test: "31.31.1", name: "multiple dots"},
		{test: "31,123.12", name: "comma and dot"},
		{test: "", name: "empty string"},
	}
	for _, test := range tests {
		s := ultrastar.NewSong()
		err := SetTag(s, TagGap, test.test)
		assert.Error(t, err, test.name)
	}
}

func TestSetTag_VideoGap(t *testing.T) {
	s := ultrastar.NewSong()
	err := SetTag(s, TagVideoGap, "123.24")
	assert.NoError(t, err)
	expected := time.Duration(123.24 * float64(time.Second))
	assert.Equal(t, expected, s.VideoGap)
}

func TestSetTag_NotesGap(t *testing.T) {
	s := ultrastar.NewSong()
	err := SetTag(s, TagNotesGap, "123")
	assert.NoError(t, err)
	assert.Equal(t, ultrastar.Beat(123), s.NotesGap)
}

// TODO: Probably more tag tests
