package txt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"codello.dev/ultrastar"
)

// TestSetTag_StringMetadata tests that known string tags are stored correctly.
func TestSetTag(t *testing.T) {
	s := ultrastar.NewSong()
	cases := []struct {
		name  string
		tag   string
		field *string
	}{
		{"TagMP3", TagMP3, &s.AudioFileName},
		{"TagVideo", TagVideo, &s.VideoFileName},
		{"TagCover", TagCover, &s.CoverFileName},
		{"TagBackground", TagBackground, &s.BackgroundFileName},

		{"TagTitle", TagTitle, &s.Title},
		{"TagArtist", TagArtist, &s.Artist},
		{"TagGenre", TagGenre, &s.Genre},
		{"TagEdition", TagEdition, &s.Edition},
		{"TagCreator", TagCreator, &s.Creator},
		{"TagAuthor", TagAuthor, &s.Creator},
		{"TagLanguage", TagLanguage, &s.Language},

		{"TagComment", TagComment, &s.Comment},
		{"TagDuetSingerP1", TagDuetSingerP1, &s.DuetSinger1},
		{"TagDuetSingerP2", TagDuetSingerP2, &s.DuetSinger2},
		{"TagP1", TagP1, &s.DuetSinger1},
		{"TagP2", TagP2, &s.DuetSinger2},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := SetTag(s, c.tag, "Hello World")
			assert.NoError(t, err)
			assert.Equal(t, "Hello World", *c.field)
			assert.Empty(t, s.CustomTags)
			// Reset field for next test
			*c.field = ""
		})
	}

	t.Run("custom tag", func(t *testing.T) {
		s := ultrastar.NewSong()
		err := SetTag(s, "My Tag", "My Value")
		assert.NoError(t, err)
		assert.Equal(t, 1, len(s.CustomTags))
		assert.NotContains(t, s.CustomTags, "My Tag")
		assert.Contains(t, s.CustomTags, "MY TAG")
		assert.Equal(t, "My Value", s.CustomTags["MY TAG"])
	})

	t.Run("case insensitive tags", func(t *testing.T) {
		s := ultrastar.NewSong()
		err := SetTag(s, "title", "Some Value")
		assert.NoError(t, err)
		assert.Equal(t, "Some Value", s.Title)
	})
}

func TestSetTag_gap(t *testing.T) {
	s := ultrastar.NewSong()
	err := SetTag(s, TagGap, "123.24")
	assert.NoError(t, err)
	expected := time.Duration(123.24 * float64(time.Millisecond))
	assert.Equal(t, expected, s.Gap)
}

func TestSetTag_invalidGap(t *testing.T) {
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

func TestSetTag_videoGap(t *testing.T) {
	s := ultrastar.NewSong()
	err := SetTag(s, TagVideoGap, "123.24")
	assert.NoError(t, err)
	expected := time.Duration(123.24 * float64(time.Second))
	assert.Equal(t, expected, s.VideoGap)
}

// TODO: Probably more tag tests
