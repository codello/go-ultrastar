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
	ErrEmptyLine         = errors.New("unexpected empty line")
	ErrInvalidPNumber    = errors.New("invalid P-number")
	ErrUnexpectedPNumber = errors.New("unexpected P number")
	ErrInvalidNote       = errors.New("invalid note")
	ErrInvalidLineBreak  = errors.New("invalid line break")
	ErrInvalidBPMChange  = errors.New("invalid BPM change")
)

type ParseError struct {
	line int
	err  error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error at line %d: %v", e.Line(), e.err)
}

func (e *ParseError) Line() int {
	return e.line
}

func (e *ParseError) Unwrap() error {
	return e.err
}

func ParseSongFromString(s string, opts ...Option) (*ultrastar.Song, error) {
	return ParseSong(strings.NewReader(s), opts...)
}

// TODO: Doc: May return partial result
func ParseSong(r io.Reader, opts ...Option) (*ultrastar.Song, error) {
	o := options{}
	for _, opt := range opts {
		opt(&o)
	}
	s := newScanner(r)
	song := ultrastar.NewSong()
	relative := false
	var bpm ultrastar.BPM
	if err := parseTags(s, song, &bpm, &relative); err != nil {
		return song, err
	}
	if err := s.skipEmptyLines(); err != nil {
		return song, &ParseError{s.line(), err}
	}
	if err := parseMusic(s, song, bpm, relative); err != nil {
		return song, err
	}
	return song, nil
}

func parseTags(s *scanner, song *ultrastar.Song, bpm *ultrastar.BPM, relative *bool) error {
	var line, tag, value string
	for s.scan() {
		line = s.text()
		if line == "" || line[0] != '#' {
			s.unScan()
			break
		}
		tag, value = splitTag(line)
		if tag == TagRelative {
			*relative = strings.ToUpper(value) == "YES"
		} else if tag == TagBPM {
			parsed, err := parseFloat(value)
			if err != nil {
				return &ParseError{s.line(), err}
			}
			*bpm = ultrastar.BPM(parsed * 4)
		} else if tag == TagEncoding {
			// Encoding tag is intentionally ignored
		} else if err := SetTag(song, tag, value); err != nil {
			return &ParseError{s.line(), err}
		}
	}

	if err := s.err(); err != nil {
		return &ParseError{s.line(), s.err()}
	}
	return nil
}

func splitTag(line string) (string, string) {
	var tag, value string
	index := strings.Index(line, ":")
	if index < 0 {
		tag, value = line[1:], ""
	} else {
		tag, value = line[1:index], line[index+1:]
	}
	return strings.ToUpper(strings.TrimSpace(tag)), strings.TrimSpace(value)
}

func parseMusic(s *scanner, song *ultrastar.Song, bpm ultrastar.BPM, relative bool) error {
	if !s.scan() {
		if s.err() != nil {
			return nil
		} else {
			return &ParseError{s.line(), s.err()}
		}
	}
	line := s.text()
	duet := line[0] == 'P'
	s.unScan()

	var player int
	rel := [2]ultrastar.Beat{0, 0}
	song.MusicP1.SetBPM(bpm)
	if duet {
		song.MusicP2 = ultrastar.NewMusicWithBPM(bpm)
	}
	music := [2]*ultrastar.Music{song.MusicP1, song.MusicP2}

LineLoop:
	for s.scan() {
		line = s.text()
		if line == "" {
			return &ParseError{s.line(), ErrEmptyLine}
		}
		switch line[0] {
		case uint8(ultrastar.NoteTypeRegular), uint8(ultrastar.NoteTypeGolden), uint8(ultrastar.NoteTypeFreestyle), uint8(ultrastar.NoteTypeRap), uint8(ultrastar.NoteTypeGoldenRap):
			note, err := ParseNote(line)
			if err != nil {
				return &ParseError{s.line(), ErrInvalidNote}
			}
			music[player].Notes = append(music[player].Notes, note)
		case '-':
			fields := strings.Fields(line[1:])
			if (!relative && len(fields) != 1) || (relative && len(fields) != 2) {
				return &ParseError{s.line(), ErrInvalidLineBreak}
			}
			beat, err := strconv.Atoi(fields[0])
			if err != nil {
				return &ParseError{s.line(), ErrInvalidLineBreak}
			}
			if relative {
				offset, err := strconv.Atoi(fields[1])
				if err != nil {
					return &ParseError{s.line(), ErrInvalidLineBreak}
				}
				rel[player] += ultrastar.Beat(offset)
			}
			music[player].LineBreaks = append(music[player].LineBreaks, rel[player]+ultrastar.Beat(beat))
		case 'P':
			if !duet {
				return &ParseError{s.line(), ErrUnexpectedPNumber}
			}
			p, err := strconv.Atoi(strings.TrimSpace(line[1:]))
			if err != nil || p < 1 || p > 2 {
				return &ParseError{s.line(), ErrInvalidPNumber}
			}
			player = p - 1
		case 'B':
			fields := strings.Fields(line[1:])
			if len(fields) != 2 {
				return &ParseError{s.line(), ErrInvalidBPMChange}
			}
			beat, err := strconv.Atoi(fields[0])
			if err != nil {
				return &ParseError{s.line(), ErrInvalidBPMChange}
			}
			bpm, err := parseFloat(fields[1])
			if err != nil {
				return &ParseError{s.line(), ErrInvalidBPMChange}
			}
			song.MusicP1.BPMs = append(song.MusicP1.BPMs, ultrastar.BPMChange{
				Start: rel[0] + ultrastar.Beat(beat),
				BPM:   ultrastar.BPM(bpm * 4),
			})
			if duet {
				// Even in duet mode BPM changes are always relative to P1
				song.MusicP2.BPMs = append(song.MusicP2.BPMs, ultrastar.BPMChange{
					Start: rel[0] + ultrastar.Beat(beat),
					BPM:   ultrastar.BPM(bpm * 4),
				})
			}
		case 'E':
			break LineLoop
		}
	}
	if err := s.err(); err != nil {
		return &ParseError{s.line(), err}
	}
	song.MusicP1.Sort()
	if duet {
		song.MusicP2.Sort()
	}
	return nil
}

// parseFloat converts a string from an UltraStar txt to a float. This function
// implements some special parsing behavior to parse UltraStar floats,
// specifically supporting a comma as decimal separator.
func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(strings.Replace(s, ",", ".", 1), 64)
}
