package txt

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"codello.dev/ultrastar"
)

// A Format defines how an [ultrastar.Song] is serialized to TXT.
// This is analogous to the [Dialect] of the parser.
//
// Methods on Format values are safe for concurrent use by multiple goroutines
// as long as the dialect value remains unchanged.
type Format struct {
	// FieldSeparator is a character used to separate fields in note line and line breaks.
	// This should only be set to a space or tab.
	//
	// Characters other than space or tab may or may not work and
	// will most likely result in invalid songs.
	FieldSeparator rune

	// Relative indicates that the writer will write music in relative mode.
	// This is a legacy format that is not recommended anymore.
	Relative bool

	// CommaFloat indicates that floating point values should use a comma as decimal separator.
	CommaFloat bool

	// TODO: Allow the format to customize the order of tags
}

// FormatDefault is the default format.
// The default format is fully compatible with all known Karaoke games.
var FormatDefault = Format{
	FieldSeparator: ' ',
	Relative:       false,
	CommaFloat:     false,
}

// FormatRelative is equal to the FormatDefault but will write songs in relative mode.
// Relative mode is basically deprecated and should only be used for good reasons.
var FormatRelative = Format{
	FieldSeparator: ' ',
	Relative:       true,
	CommaFloat:     false,
}

// WriteSong serializes s into w.
// This is a convenience method for [Format.WriteSong].
func WriteSong(w io.Writer, s *ultrastar.Song) error {
	return FormatDefault.WriteSong(w, s)
}

// allTags are all tag values that have a corresponding field in [ultrastar.Song].
// The order of this slice determines the order of tags in TXT files.
var allTags = []string{
	TagTitle, TagArtist, TagLanguage, TagEdition, TagGenre, TagYear,
	TagCreator, TagComment, TagMP3, TagCover, TagBackground, TagVideo,
	TagVideoGap, TagStart, TagEnd, TagPreviewStart, TagMedleyStartBeat,
	TagMedleyEndBeat, TagCalcMedley, TagBPM, TagGap, TagP1, TagP2,
}

// WriteSong writes the song s to w in the UltraStar txt format.
// If an error occurs it is returned, otherwise nil is returned.
func (f Format) WriteSong(w io.Writer, s *ultrastar.Song) error {
	for _, tag := range allTags {
		value := f.GetTag(s, tag)
		if value != "" {
			if err := f.WriteTag(w, tag, value); err != nil {
				return err
			}
		}
	}
	if f.Relative {
		if err := f.WriteTag(w, TagRelative, "YES"); err != nil {
			return err
		}
	}
	for tag, value := range s.CustomTags {
		if err := f.WriteTag(w, tag, value); err != nil {
			return err
		}
	}
	if s.IsDuet() {
		if _, err := io.WriteString(w, "P1\n"); err != nil {
			return err
		}
	}
	if err := f.WriteNotes(w, s.NotesP1); err != nil {
		return err
	}
	if s.IsDuet() {
		if _, err := io.WriteString(w, "P2\n"); err != nil {
			return err
		}
		if err := f.WriteNotes(w, s.NotesP2); err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, "E\n")
	return err
}

// WriteTag writes a single tag.
// Neither the tag nor the value are validated or normalized, they are written as-is.
func (f Format) WriteTag(w io.Writer, tag string, value string) error {
	s := fmt.Sprintf("#%s:%s\n", tag, value)
	_, err := io.WriteString(w, s)
	return err
}

// WriteNotes writes all notes, line breaks and BPM changes in m in standard UltraStar format.
//
// Depending on the value of f.Relative the music may be written in relative mode.
// A #RELATIVE tag is NOT written automatically in this case.
func (f Format) WriteNotes(w io.Writer, ns ultrastar.Notes) error {
	rel := ultrastar.Beat(0)
	for _, n := range ns {
		if err := f.WriteNoteRel(w, n, &rel); err != nil {
			return err
		}
	}
	return nil
}

// WriteNote writes a single note line.
// The note is written as-is, even if w is in relative mode.
func (f Format) WriteNote(w io.Writer, n ultrastar.Note) error {
	return f.WriteNoteRel(w, n, nil)
}

// WriteNoteRel writes a single note.
// If f.Relative is true, the note start is adjusted by rel.
// If n is a line break, rel is updated accordingly.
func (f Format) WriteNoteRel(w io.Writer, n ultrastar.Note, rel *ultrastar.Beat) error {
	var parts []string
	if f.Relative && rel != nil {
		n.Start -= *rel
	}
	if n.Type.IsLineBreak() {
		beat := strconv.Itoa(int(n.Start))
		if f.Relative {
			parts = []string{string(ultrastar.NoteTypeLineBreak), beat, beat}
			if rel != nil {
				*rel += n.Start
			}
		} else {
			parts = []string{string(ultrastar.NoteTypeLineBreak), beat}
		}
	} else {
		parts = []string{
			string(n.Type),
			strconv.Itoa(int(n.Start)),
			strconv.Itoa(int(n.Duration)),
			strconv.Itoa(int(n.Pitch)),
			n.Text,
		}
	}
	s := strings.Join(parts, string(f.FieldSeparator)) + "\n"
	_, err := io.WriteString(w, s)
	return err
}
