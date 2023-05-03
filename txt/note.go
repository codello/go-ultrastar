package txt

import (
	"errors"
	"fmt"
	"strconv"
	"unicode"

	"github.com/codello/ultrastar"
)

// ParseNote converts an UltraStar-style note line into a [Note] instance.
// The format is as follows: `X A B C Text` where
//   - `X` is a single character denoting the [NoteType]. The note type must be
//     valid as determined by [ultrastar.NoteType.IsValid].
//   - `A` is an integer denoting the start beat of the note
//   - `B` is an integer denoting the duration of the note
//   - `C` is an integer denoting the pitch of the note
//   - `Text` is the remaining text on the line denoting the text of the syllable.
//
// If an error occurs the returned note may be partially initialized. However,
// this behavior should not be relied upon.
func ParseNote(s string) (ultrastar.Note, error) {
	n := ultrastar.Note{}
	if s == "" {
		return n, errors.New("invalid note type")
	}
	runes := []rune(s[0:])
	nType := ultrastar.NoteType(runes[0])
	runes = runes[1:]
	if !nType.IsValid() {
		return n, fmt.Errorf("invalid note type: %c", nType)
	}
	n.Type = nType

	value, runes := nextField(runes)
	start, err := strconv.Atoi(value)
	n.Start = ultrastar.Beat(start)
	if err != nil {
		return n, fmt.Errorf("invalid note start: %w", err)
	}

	value, runes = nextField(runes)
	duration, err := strconv.Atoi(value)
	n.Duration = ultrastar.Beat(duration)
	if err != nil {
		return n, fmt.Errorf("invalid note duration: %w", err)
	}

	value, runes = nextField(runes)
	pitch, err := strconv.Atoi(value)
	n.Pitch = ultrastar.Pitch(pitch)
	if err != nil {
		return n, fmt.Errorf("invalid note pitch: %w", err)
	}

	if len(runes) == 0 {
		return n, errors.New("empty note text")
	}
	if !unicode.Is(unicode.White_Space, runes[0]) {
		return n, errors.New("missing whitespace after note pitch")
	}
	if len(runes) < 2 {
		return n, errors.New("empty note text")
	}
	n.Text = string(runes[1:])
	return n, nil
}

// nextField finds the next whitespace-separated field in a slice of runes. The
// function skips over leading whitespace and finds a consecutive run of
// non-whitespace runes. Returned is the field as string and the remaining runes
// as a slice.
func nextField(runes []rune) (string, []rune) {
	start := 0
	for ; start < len(runes); start++ {
		if !unicode.Is(unicode.White_Space, runes[start]) {
			break
		}
	}
	end := start
	for ; end < len(runes); end++ {
		if unicode.Is(unicode.White_Space, runes[end]) {
			break
		}
	}
	return string(runes[start:end]), runes[end:]
}
