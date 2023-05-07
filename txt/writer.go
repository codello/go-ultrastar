package txt

import (
	"errors"
	"fmt"
	"github.com/codello/ultrastar"
	"io"
	"strconv"
	"strings"
)

var (
	ErrBPMMismatch = errors.New("duet voices have different BPMs")

	allTags = []string{
		TagTitle, TagArtist, TagLanguage, TagEdition, TagGenre, TagYear,
		TagCreator, TagMP3, TagCover, TagBackground, TagVideo, TagVideoGap,
		TagResolution, TagNotesGap, TagStart, TagEnd, TagPreviewStart,
		TagMedleyStartBeat, TagMedleyEndBeat, TagBPM, TagGap, TagP1, TagP2,
	}
)

type Format struct {
	FieldSeparator rune
	Relative       bool
}

var (
	FieldSeparatorSpace = ' '
	FieldSeparatorTab   = '\t'

	defaultFormat Format = Format{
		FieldSeparator: FieldSeparatorTab,
		Relative:       false,
	}
)

func WriteSongToString(s *ultrastar.Song) (string, error) {
	b := &strings.Builder{}
	if err := WriteSong(b, s); err != nil {
		return "", err
	}
	return b.String(), nil
}

func WriteSong(w io.Writer, s *ultrastar.Song) error {
	return defaultFormat.WriteSong(w, s)
}

func (f *Format) WriteSong(w io.Writer, s *ultrastar.Song) error {
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
	for _, tag := range allTags {
		value := GetTag(s, tag)
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
	m := &ultrastar.Music{
		Notes:      s.MusicP1.Notes,
		LineBreaks: s.MusicP1.LineBreaks,
	}
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
		m.LineBreaks = s.MusicP2.LineBreaks
		m.BPMs = nil
		if err := f.WriteMusic(w, m); err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, "E\n")
	return err
}

func WriteTag(w io.Writer, tag string, value string) error {
	return defaultFormat.WriteTag(w, tag, value)
}

func (f *Format) WriteTag(w io.Writer, tag string, value string) error {
	s := fmt.Sprintf("#%s:%s\n", tag, value)
	_, err := io.WriteString(w, s)
	return err
}

func WriteMusic(w io.Writer, m *ultrastar.Music) error {
	return defaultFormat.WriteMusic(w, m)
}

func (f *Format) WriteMusic(w io.Writer, m *ultrastar.Music) error {
	var i, j, k int
	rel := ultrastar.Beat(0)
	noteBeat, lineBreakBeat, bpmBeat := ultrastar.MaxBeat, ultrastar.MaxBeat, ultrastar.MaxBeat
	if i < len(m.Notes) {
		noteBeat = m.Notes[i].Start
	}
	if j < len(m.LineBreaks) {
		lineBreakBeat = m.LineBreaks[j]
	}
	if k < len(m.BPMs) {
		bpmBeat = m.BPMs[k].Start
	}
	for i < len(m.Notes) || j < len(m.LineBreaks) || k < len(m.BPMs) {
		if noteBeat < lineBreakBeat && noteBeat < bpmBeat {
			n := m.Notes[i]
			n.Start -= rel
			if err := f.WriteNote(w, n); err != nil {
				return err
			}
			i++
			noteBeat = ultrastar.MaxBeat
			if i < len(m.Notes) {
				noteBeat = m.Notes[i].Start
			}
		} else if lineBreakBeat < bpmBeat {
			beat := m.LineBreaks[j]
			beat -= rel
			if err := f.writeLineBreak(w, beat, &rel); err != nil {
				return err
			}
			j++
			lineBreakBeat = ultrastar.MaxBeat
			if j < len(m.LineBreaks) {
				lineBreakBeat = m.LineBreaks[j]
			}
		} else {
			c := m.BPMs[k]
			c.Start -= rel
			if err := f.writeBPMChange(w, c); err != nil {
				return err
			}
			k++
			bpmBeat = ultrastar.MaxBeat
			if k < len(m.BPMs) {
				bpmBeat = m.BPMs[k].Start
			}
		}
	}
	return nil
}

func WriteNote(w io.Writer, n ultrastar.Note) error {
	return defaultFormat.WriteNote(w, n)
}

func (f *Format) WriteNote(w io.Writer, n ultrastar.Note) error {
	parts := []string{
		string(n.Type),
		strconv.Itoa(int(n.Start)),
		strconv.Itoa(int(n.Duration)),
		strconv.Itoa(int(n.Pitch)),
		n.Text,
	}
	s := strings.Join(parts, string(f.FieldSeparator)) + "\n"
	_, err := io.WriteString(w, s)
	return err
}

func (f *Format) writeLineBreak(w io.Writer, b ultrastar.Beat, rel *ultrastar.Beat) error {
	beatString := strconv.Itoa(int(b))
	var s string
	if f.Relative {
		s = strings.Join([]string{"-", beatString, beatString}, string(f.FieldSeparator)) + "\n"
		*rel += b
	} else {
		s = "-" + string(f.FieldSeparator) + beatString + "\n"
	}
	_, err := io.WriteString(w, s)
	return err
}

func (f *Format) writeBPMChange(w io.Writer, c ultrastar.BPMChange) error {
	parts := []string{
		"B",
		strconv.Itoa(int(c.Start)),
		strconv.FormatFloat(float64(c.BPM), 'f', -1, 64),
	}
	s := strings.Join(parts, string(f.FieldSeparator)) + "\n"
	_, err := io.WriteString(w, s)
	return err
}
