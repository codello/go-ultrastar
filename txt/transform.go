package txt

import (
	"fmt"
	"strings"

	"golang.org/x/text/transform"

	"codello.dev/ultrastar"
)

// A TransformError indicates an error that occurred during applying a [transform.Transformer] to an [ultrastar.Song].
// Depending on the function used different fields may be nil or an empty map.
type TransformError struct {
	// errors occurred when transforming tag values. The key is the tag name for
	// which the value could not be transformed.
	TagErrors map[string]error
	// errors occurred when transforming custom tag keys. The key is the (untransformed)
	// tag name that could not be transformed.
	TagKeyErrors map[string]error
	// errors occurred when transforming note texts. The key is the note index
	// that caused the error.
	NoteErrors map[int]error
}

// Error implements the error interface.
func (e *TransformError) Error() string {
	errs := make([]string, 0, len(e.TagErrors)+len(e.TagKeyErrors)+len(e.NoteErrors))
	for key, err := range e.TagErrors {
		errs = append(errs, fmt.Sprintf("transform error for tag %s: %s", key, err.Error()))
	}
	for key, err := range e.TagKeyErrors {
		errs = append(errs, fmt.Sprintf("transform error for tag key %s: %s", key, err.Error()))
	}
	for idx, err := range e.NoteErrors {
		errs = append(errs, fmt.Sprintf("transform error for note %d: %s", idx, err.Error()))
	}
	return strings.Join(errs, "\n")
}

// Unwrap implements the error interface.
func (e *TransformError) Unwrap() []error {
	errs := make([]error, 0, len(e.TagErrors)+len(e.TagKeyErrors)+len(e.NoteErrors))
	i := 0
	for _, err := range e.TagErrors {
		errs[i] = err
		i++
	}
	for _, err := range e.TagKeyErrors {
		errs[i] = err
		i++
	}
	for _, err := range e.NoteErrors {
		errs[i] = err
		i++
	}
	return errs
}

// TransformSong applies the given [transform.Transformer] to all texts in s.
// All texts include tag values, custom tag names, and all note texts.
//
// If an error occurs during the transformation of a tag or note
// the error will be recorded and the tag or note will not be modified.
// The transformation however will continue through the entire song.
// The error returned is a TransformError that allows you to inspect the places that caused errors.
func TransformSong(s *ultrastar.Song, t transform.Transformer) error {
	tErr := &TransformError{
		TagErrors:    map[string]error{},
		TagKeyErrors: map[string]error{},
		NoteErrors:   nil,
	}

	transformTagValue(t, &s.AudioFile, TagMP3, tErr)
	transformTagValue(t, &s.VideoFile, TagVideo, tErr)
	transformTagValue(t, &s.CoverFile, TagCover, tErr)
	transformTagValue(t, &s.BackgroundFile, TagBackground, tErr)

	transformTagValue(t, &s.Title, TagTitle, tErr)
	transformTagValue(t, &s.Artist, TagArtist, tErr)
	transformTagValue(t, &s.Genre, TagGenre, tErr)
	transformTagValue(t, &s.Edition, TagEdition, tErr)
	transformTagValue(t, &s.Creator, TagCreator, tErr)
	transformTagValue(t, &s.Language, TagLanguage, tErr)
	transformTagValue(t, &s.Comment, TagComment, tErr)
	transformTagValue(t, &s.DuetSinger1, TagDuetSingerP1, tErr)
	transformTagValue(t, &s.DuetSinger2, TagDuetSingerP2, tErr)

	newCustomTags := make(map[string]string, len(s.CustomTags))
	for key, value := range s.CustomTags {
		transformTagKey(t, &key, tErr)
		transformTagValue(t, &value, key, tErr)
		newCustomTags[key] = value
	}
	s.CustomTags = newCustomTags

	if err := TransformMusic(s.MusicP1, t); err != nil {
		tErr.NoteErrors = err.(*TransformError).NoteErrors
	}
	if err := TransformMusic(s.MusicP2, t); err != nil {
		if tErr.NoteErrors == nil {
			tErr.NoteErrors = err.(*TransformError).NoteErrors
		} else {
			for key, value := range err.(*TransformError).NoteErrors {
				tErr.NoteErrors[key] = value
			}
		}
	}
	if len(tErr.TagErrors) > 0 || len(tErr.TagKeyErrors) > 0 || len(tErr.NoteErrors) > 0 {
		return tErr
	}
	return nil
}

// transformTagKey applies t to *v and stores the result in *v if the transformation was successful.
// If an error occurs, it will be appended to err.TagKeyErrors.
func transformTagKey(t transform.Transformer, v *string, err *TransformError) {
	v2, _, err2 := transform.String(t, *v)
	if err2 != nil {
		err.TagKeyErrors[*v] = err2
	} else {
		*v = v2
	}
}

// transformTagValue applies t to *v and stores the result in *v if the transformation was successful.
// If an error occurs, it will be appended to err.TagErrors[tag].
func transformTagValue(t transform.Transformer, v *string, tag string, err *TransformError) {
	v2, _, err2 := transform.String(t, *v)
	if err2 != nil {
		err.TagErrors[tag] = err2
	} else {
		*v = v2
	}
}

// TransformMusic applies t to the text of every note in m.
// If an error occurs the return value will be of type TransformError and have its NoteErrors field set.
// The note text that caused the error will remain unchanged.
// Even if an error occurs the remaining music will still be transformed.
func TransformMusic(m *ultrastar.Music, t transform.Transformer) error {
	if m == nil {
		return nil
	}
	errs := map[int]error{}
	for i := range m.Notes {
		if m.Notes[i].Type.IsLineBreak() {
			continue
		}
		s, _, err := transform.String(t, m.Notes[i].Text)
		if err != nil {
			errs[i] = err
		}
		m.Notes[i].Text = s
	}
	if len(errs) > 0 {
		return &TransformError{NoteErrors: errs}
	}
	return nil
}
