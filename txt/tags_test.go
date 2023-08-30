package txt

import (
	"testing"
	"time"

	"codello.dev/ultrastar"
)

// TestSetTag_StringMetadata tests that known string tags are stored correctly.
func TestSetTag(t *testing.T) {
	s := ultrastar.Song{}
	cases := map[string]struct {
		tag   string
		field *string
	}{
		"TagMP3":        {TagMP3, &s.AudioFileName},
		"TagVideo":      {TagVideo, &s.VideoFileName},
		"TagCover":      {TagCover, &s.CoverFileName},
		"TagBackground": {TagBackground, &s.BackgroundFileName},

		"TagTitle":    {TagTitle, &s.Title},
		"TagArtist":   {TagArtist, &s.Artist},
		"TagGenre":    {TagGenre, &s.Genre},
		"TagEdition":  {TagEdition, &s.Edition},
		"TagCreator":  {TagCreator, &s.Creator},
		"TagAuthor":   {TagAuthor, &s.Creator},
		"TagLanguage": {TagLanguage, &s.Language},

		"TagComment":      {TagComment, &s.Comment},
		"TagDuetSingerP1": {TagDuetSingerP1, &s.DuetSinger1},
		"TagDuetSingerP2": {TagDuetSingerP2, &s.DuetSinger2},
		"TagP1":           {TagP1, &s.DuetSinger1},
		"TagP2":           {TagP2, &s.DuetSinger2},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			err := SetTag(&s, c.tag, "Hello World")
			if err != nil {
				t.Errorf("SetTag(&s, %q, %q) caused an unexpected error: %s", c.tag, "Hello World", err)
			}
			if *c.field != "Hello World" {
				t.Errorf("SetTag(&s, %q, %q) did not modify field at %p", c.tag, "Hello World", c.field)
			}
			if len(s.CustomTags) > 0 {
				t.Errorf("SetTag(&s, %q, %q) updated a custom field, expected known field", *c.field, "Hello World")
			}
			// Reset field for next test
			*c.field = ""
		})
	}

	t.Run("custom tag", func(t *testing.T) {
		s := ultrastar.Song{}
		err := SetTag(&s, "My Tag", "My Value")
		if err != nil {
			t.Errorf("SetTag(&s, %q, %q) caused an unexpected error: %s", "My Tag", "My Value", err)
		}
		if len(s.CustomTags) != 1 {
			t.Fatalf("SetTag(&s, %q, %q) set %d tags, expected 1", "My Tag", "My Value", len(s.CustomTags))
		}
		val, ok := s.CustomTags["MY TAG"]
		if !ok {
			t.Errorf("SetTag(&s, %q, %q) did not set custom tag %q", "My Tag", "My Value", "MY TAG")
		}
		if val != "My Value" {
			t.Errorf("SetTag(&s, %q, %q) set a value of %q, expected %q", "My Tag", "My Value", val, "My Value")
		}
	})

	t.Run("case insensitive tags", func(t *testing.T) {
		s := ultrastar.Song{}
		err := SetTag(&s, "title", "Some Value")
		if err != nil {
			t.Errorf("SetTag(&s, %q, %q) caused an unexpected error: %s", "title", "Some Value", err)
		}
		if s.Title != "Some Value" {
			t.Errorf("SetTag(&s, %q, %q) set s.Title to %q, expected %q", "title", "Some Value", s.Title, "Some Value")
		}
	})

	t.Run("gap", func(t *testing.T) {
		s := ultrastar.Song{}
		err := SetTag(&s, TagGap, "123.24")
		if err != nil {
			t.Errorf("SetTag(&s, %q, %q) caused an unexpected error: %s", TagGap, "123.24", err)
		}
		expected := time.Duration(123.24 * float64(time.Millisecond))
		if s.Gap != expected {
			t.Errorf("SetTag(&s, %q, %q) set s.Tag to %s, expected %s", TagGap, "123.24", s.Gap, expected)
		}
	})

	t.Run("invalid gap", func(t *testing.T) {
		tests := map[string]string{
			"letters":       "31abc",
			"multiple dots": "31.31.1",
			"comma and dot": "31,123.12",
			"empty string":  "",
		}
		for name, test := range tests {
			t.Run(name, func(t *testing.T) {
				s := ultrastar.Song{}
				err := SetTag(&s, TagGap, test)
				if err == nil {
					t.Errorf("SetTag(&s, %q, %q) did not cause an error, expected one", TagGap, test)
				}
			})
		}
	})

	t.Run("video gap", func(t *testing.T) {
		s := ultrastar.Song{}
		err := SetTag(&s, TagVideoGap, "123.24")
		if err != nil {
			t.Errorf("SetTag(&s, %q, %q) caused an unexpected error: %s", TagVideoGap, "123.24", err)
		}
		expected := time.Duration(123.24 * float64(time.Second))
		if s.VideoGap != expected {
			t.Errorf("SetTag(&s, %q, %q) set s.VideoGap to %s, expected %s", TagVideoGap, "123.24", s.VideoGap, expected)
		}
	})
}

// TODO: Probably more tag tests
