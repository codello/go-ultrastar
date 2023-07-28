package txt

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"codello.dev/ultrastar"
)

// These errors can occur while writing a song to TXT format.
var (
	// ErrBPMMismatch indicates that the different voices of a duet have different BPMs.
	// The UltraStar TXT format does not support this scenario.
	ErrBPMMismatch = errors.New("duet voices have different BPMs")
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
var FormatDefault = &Format{
	FieldSeparator: ' ',
	Relative:       false,
	CommaFloat:     false,
}

// FormatRelative is equal to the FormatDefault but will write songs in relative mode.
// Relative mode is basically deprecated and should only be used for good reasons.
var FormatRelative = &Format{
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
func (f *Format) WriteSong(w io.Writer, s *ultrastar.Song) error {
	for _, tag := range allTags {
		if tag == TagBPM {
			if s.IsDuet() {
				if len(s.MusicP1.BPMs) != len(s.MusicP2.BPMs) {
					return ErrBPMMismatch
				}
				for i, b := range s.MusicP1.BPMs {
					if b != s.MusicP2.BPMs[i] {
						return ErrBPMMismatch
					}
				}
			}
		}
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
	m := &ultrastar.Music{Notes: s.MusicP1.Notes}
	if len(s.MusicP1.BPMs) > 0 {
		m.BPMs = s.MusicP1.BPMs[1:]
	}
	if err := f.WriteMusic(w, m); err != nil {
		return err
	}
	if s.IsDuet() {
		if _, err := io.WriteString(w, "P2\n"); err != nil {
			return err
		}
		m.Notes = s.MusicP2.Notes
		m.BPMs = nil
		if err := f.WriteMusic(w, m); err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, "E\n")
	return err
}

// WriteTag writes a single tag.
// Neither the tag nor the value are validated or normalized, they are written as-is.
func (f *Format) WriteTag(w io.Writer, tag string, value string) error {
	s := fmt.Sprintf("#%s:%s\n", tag, value)
	_, err := io.WriteString(w, s)
	return err
}

// WriteMusic writes all notes, line breaks and BPM changes in m in standard UltraStar format.
//
// Depending on the value of f.Relative the music may be written in relative mode.
// A #RELATIVE tag is NOT written automatically in this case.
func (f *Format) WriteMusic(w io.Writer, m *ultrastar.Music) error {
	var i, j int
	rel := ultrastar.Beat(0)
	noteBeat, bpmBeat := ultrastar.MaxBeat, ultrastar.MaxBeat
	if i < len(m.Notes) {
		noteBeat = m.Notes[i].Start
	}
	if j < len(m.BPMs) {
		bpmBeat = m.BPMs[j].Start
	}
	for i < len(m.Notes) || j < len(m.BPMs) {
		if noteBeat < bpmBeat {
			n := m.Notes[i]
			if err := f.WriteNoteRel(w, n, &rel); err != nil {
				return err
			}
			i++
			noteBeat = ultrastar.MaxBeat
			if i < len(m.Notes) {
				noteBeat = m.Notes[i].Start
			}
		} else {
			c := m.BPMs[j]
			if err := f.writeBPMChange(w, c, rel); err != nil {
				return err
			}
			j++
			bpmBeat = ultrastar.MaxBeat
			if j < len(m.BPMs) {
				bpmBeat = m.BPMs[j].Start
			}
		}
	}
	return nil
}

// WriteNote writes a single note line.
// The note is written as-is, even if w is in relative mode.
func (f *Format) WriteNote(w io.Writer, n ultrastar.Note) error {
	return f.WriteNoteRel(w, n, nil)
}

// WriteNoteRel writes a single note.
// If f.Relative is true, the note start is adjusted by rel.
// If n is a line break, rel is updated accordingly.
func (f *Format) WriteNoteRel(w io.Writer, n ultrastar.Note, rel *ultrastar.Beat) error {
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

// writeBPMChange writes a BPM change line.
func (f *Format) writeBPMChange(w io.Writer, c ultrastar.BPMChange, rel ultrastar.Beat) error {
	if f.Relative {
		c.Start -= rel
	}
	parts := []string{
		"B",
		strconv.Itoa(int(c.Start)),
		strconv.FormatFloat(float64(c.BPM), 'f', -1, 64),
	}
	s := strings.Join(parts, string(f.FieldSeparator)) + "\n"
	_, err := io.WriteString(w, s)
	return err
}
