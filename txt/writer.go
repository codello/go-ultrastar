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

const (
	FieldSeparatorSpace rune = ' '
	FieldSeparatorTab   rune = '\t'
)

type Writer struct {
	w io.Writer

	FieldSeparator rune
	Relative       bool
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w:              w,
		FieldSeparator: FieldSeparatorTab,
		Relative:       false,
	}
}

func (w *Writer) WriteSong(s *ultrastar.Song) error {
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
		if _, err := io.WriteString(w.w, "P1\n"); err != nil {
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
	if err := w.WriteMusic(m); err != nil {
		return err
	}
	if s.IsDuet() {
		if _, err := io.WriteString(w.w, "P2\n"); err != nil {
			return err
		}
		m.Notes = s.MusicP2.Notes
		m.LineBreaks = s.MusicP2.LineBreaks
		m.BPMs = nil
		if err := w.WriteMusic(m); err != nil {
			return err
		}
	}
	_, err := io.WriteString(w.w, "E\n")
	return err
}

func (w *Writer) WriteTag(tag string, value string) error {
	s := fmt.Sprintf("#%s:%s\n", tag, value)
	_, err := io.WriteString(w.w, s)
	return err
}

func (w *Writer) WriteMusic(m *ultrastar.Music) error {
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
			if err := w.WriteNote(n); err != nil {
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
			if err := w.writeLineBreak(beat, &rel); err != nil {
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
			if err := w.writeBPMChange(c); err != nil {
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

func (w *Writer) WriteNote(n ultrastar.Note) error {
	parts := []string{
		string(n.Type),
		strconv.Itoa(int(n.Start)),
		strconv.Itoa(int(n.Duration)),
		strconv.Itoa(int(n.Pitch)),
		n.Text,
	}
	s := strings.Join(parts, string(w.FieldSeparator)) + "\n"
	_, err := io.WriteString(w.w, s)
	return err
}

func (w *Writer) writeLineBreak(b ultrastar.Beat, rel *ultrastar.Beat) error {
	beatString := strconv.Itoa(int(b))
	var s string
	if w.Relative {
		s = strings.Join([]string{"-", beatString, beatString}, string(w.FieldSeparator)) + "\n"
		*rel += b
	} else {
		s = "-" + string(w.FieldSeparator) + beatString + "\n"
	}
	_, err := io.WriteString(w.w, s)
	return err
}

func (w *Writer) writeBPMChange(c ultrastar.BPMChange) error {
	parts := []string{
		"B",
		strconv.Itoa(int(c.Start)),
		strconv.FormatFloat(float64(c.BPM), 'f', -1, 64),
	}
	s := strings.Join(parts, string(w.FieldSeparator)) + "\n"
	_, err := io.WriteString(w.w, s)
	return err
}
