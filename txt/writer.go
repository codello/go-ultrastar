package txt

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"codello.dev/ultrastar"
)

// WriteSong serializes s into w.
// This is a convenience method for [Format.WriteSong].
func WriteSong(w io.Writer, s ultrastar.Song) error {
	return NewWriter(w).WriteSong(s)
}

// A Writer implements serialization of [ultrastar.Song] serialized to TXT.
type Writer struct {
	// FieldSeparator is a character used to separate fields in note line and line breaks.
	// This should only be set to a space or tab.
	//
	// Characters other than space or tab may or may not work and
	// will most likely result in invalid songs.
	FieldSeparator rune

	// Relative indicates that the writer will write notes in relative mode.
	// This is a legacy format that is not recommended anymore.
	Relative bool

	// CommaFloat indicates that floating point values should use a comma as decimal separator.
	CommaFloat bool

	// TODO: Allow customization the order of tags

	wr  io.Writer      // underlying writer
	rel ultrastar.Beat // current relative offset
}

// NewWriter creates a new writer for UltraStar songs.
// The default settings aim to be compatible with most Karaoke games.
func NewWriter(wr io.Writer) *Writer {
	w := &Writer{
		FieldSeparator: ' ',
		Relative:       false,
		CommaFloat:     false,
	}
	w.Reset(wr)
	return w
}

// Reset configures w to be reused, writing to wr.
// This method keeps the current writer's configuration.
func (w *Writer) Reset(wr io.Writer) {
	w.wr = wr
	w.rel = 0
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
func (w *Writer) WriteSong(s ultrastar.Song) error {
	for _, tag := range allTags {
		value := getTag(s, tag, w.CommaFloat)
		if value != "" {
			if err := w.WriteTag(tag, value); err != nil {
				return err
			}
		}
	}
	if w.Relative {
		if err := w.WriteTag(TagRelative, "YES"); err != nil {
			return err
		}
	}
	for tag, value := range s.CustomTags {
		if err := w.WriteTag(tag, value); err != nil {
			return err
		}
	}
	if s.IsDuet() {
		if _, err := io.WriteString(w.wr, "P1\n"); err != nil {
			return err
		}
	}
	if err := w.WriteNotes(s.NotesP1); err != nil {
		return err
	}
	if s.IsDuet() {
		w.rel = 0
		if _, err := io.WriteString(w.wr, "P2\n"); err != nil {
			return err
		}
		if err := w.WriteNotes(s.NotesP2); err != nil {
			return err
		}
	}
	_, err := io.WriteString(w.wr, "E\n")
	return err
}

// WriteTag writes a single tag.
// Neither the tag nor the value are validated or normalized, they are written as-is.
func (w *Writer) WriteTag(tag string, value string) error {
	s := fmt.Sprintf("#%s:%s\n", tag, value)
	_, err := io.WriteString(w.wr, s)
	return err
}

// WriteNotes writes all notes, line breaks and BPM changes in m in standard UltraStar format.
//
// Depending on the value of w.Relative the notes may be written in relative mode.
// A #RELATIVE tag is NOT written automatically in this case.
func (w *Writer) WriteNotes(ns ultrastar.Notes) error {
	for _, n := range ns {
		if err := w.WriteNote(n); err != nil {
			return err
		}
	}
	return nil
}

// WriteNote writes a single note line.
// Depending on w.Relative the note is adjusted to the current relative offset.
func (w *Writer) WriteNote(n ultrastar.Note) error {
	var parts []string
	if w.Relative {
		n.Start -= w.rel
	}
	if n.Type.IsLineBreak() {
		beat := strconv.Itoa(int(n.Start))
		if w.Relative {
			parts = []string{string(ultrastar.NoteTypeLineBreak), beat, beat}
			w.rel += n.Start
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
	s := strings.Join(parts, string(w.FieldSeparator)) + "\n"
	_, err := io.WriteString(w.wr, s)
	return err
}
