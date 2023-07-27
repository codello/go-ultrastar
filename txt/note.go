package txt

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"codello.dev/ultrastar"
)

// ParseNote parses s into an [ultrastar.Note].
// This is a convenience function for [Dialect.ParseNoteRelative].
func ParseNote(s string) (ultrastar.Note, error) {
	return DialectDefault.ParseNoteRelative(s, false)
}

// ParseNote parses s into an [ultrastar.Note].
// This is a convenience function for [Dialect.ParseNoteRelative].
func (d *Dialect) ParseNote(s string) (ultrastar.Note, error) {
	return d.ParseNoteRelative(s, false)
}

// ParseNoteRelative converts an UltraStar-style note line into an ultrastar.Note
// instance. There are two formats:
//
// Regular notes follow the format "X A B C Text" where
//   - X is a single character denoting the [ultrastar.NoteType].
//     The note type must be valid as determined by [ultrastar.NoteType.IsValid].
//   - A is an integer denoting the start beat of the note
//   - B is an integer denoting the duration of the note
//   - C is an integer denoting the pitch of the note
//   - Text is the remaining text on the line denoting the text of the syllable.
//
// Line breaks follow the format "X A" or "X A B" where
//   - X is the character '-' (dash)
//   - A is an integer denoting the start beat of the line break
//   - B is an integer denoting the relative offset of the next line.
//     This format is only used when relative is true.
//
// If an error occurs the returned note may be partially initialized. However,
// this behavior should not be relied upon.
func (d *Dialect) ParseNoteRelative(s string, relative bool) (ultrastar.Note, error) {
	n := ultrastar.Note{}
	if s == "" {
		return n, errors.New("invalid note type")
	}
	nType := ultrastar.NoteType(s[0])
	s = s[1:]
	if !nType.IsValid() {
		return n, fmt.Errorf("invalid note type: %c", nType)
	}
	n.Type = nType
	if n.Type.IsLineBreak() {
		n.Text = "\n"
	}

	value, s := nextField(s)
	start, err := strconv.Atoi(value)
	n.Start = ultrastar.Beat(start)
	if err != nil {
		return n, fmt.Errorf("invalid note start: %w", err)
	}

	if nType.IsLineBreak() && !relative {
		if d.StrictLineBreaks && strings.TrimSpace(s) != "" {
			return n, fmt.Errorf("invalid line break: extra text")
		}
		return n, nil
	}

	value, s = nextField(s)
	duration, err := strconv.Atoi(value)
	n.Duration = ultrastar.Beat(duration)
	if n.Type.IsLineBreak() {
		if err != nil {
			return n, fmt.Errorf("invalid line break: invalid relative spec: %w", err)
		}
		if d.StrictLineBreaks && strings.TrimSpace(s) != "" {
			return n, fmt.Errorf("invalid line break: extra text")
		}
		return n, nil
	}
	if err != nil {
		return n, fmt.Errorf("invalid note duration: %w", err)
	}

	value, s = nextField(s)
	pitch, err := strconv.Atoi(value)
	n.Pitch = ultrastar.Pitch(pitch)
	if err != nil {
		return n, fmt.Errorf("invalid note pitch: %w", err)
	}

	if s == "" {
		return n, errors.New("empty note text")
	}
	if s[0] != ' ' && s[0] != '\t' {
		return n, errors.New("missing whitespace after note pitch")
	}
	if len(s) < 2 {
		return n, errors.New("empty note text")
	}
	n.Text = s[1:]
	return n, nil
}

// nextField finds the next whitespace-separated field in a string. The function
// skips over leading whitespace and finds a consecutive run of non-space and
// non-tab characters. Returned is the found field and the remaining string.
func nextField(s string) (string, string) {
	start := 0
	for ; start < len(s); start++ {
		if s[start] != ' ' && s[start] != '\t' {
			break
		}
	}
	end := start
	for ; end < len(s); end++ {
		if s[end] == ' ' || s[end] == '\t' {
			break
		}
	}
	return s[start:end], s[end:]
}
