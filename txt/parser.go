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
	ErrUnknownEvent      = errors.New("invalid event")
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

// Currently we have no options

func ParseSong(s string) (*ultrastar.Song, error) {
	p := NewParser(strings.NewReader(s))
	return p.ParseSong()
}

type Parser struct {
	scanner *scanner

	relative bool
	bpm      ultrastar.BPM
}

func NewParser(r io.Reader) *Parser {
	return &Parser{
		scanner: newScanner(r),
	}
}

// TODO: Doc: May return partial result
func (p *Parser) ParseSong() (*ultrastar.Song, error) {
	song := ultrastar.NewSong()
	if err := p.parseTags(song); err != nil {
		return song, p.error(err)
	}
	if err := p.scanner.skipEmptyLines(); err != nil {
		return song, p.error(err)
	}
	if err := p.parseMusic(song); err != nil {
		return song, p.error(err)
	}
	return song, nil
}

func (p *Parser) error(err error) *ParseError {
	if err == nil {
		return nil
	}
	return &ParseError{p.scanner.line(), err}
}

func (p *Parser) parseTags(song *ultrastar.Song) error {
	var line, tag, value string
	for p.scanner.scan() {
		line = p.scanner.text()
		if line == "" || line[0] != '#' {
			p.scanner.unScan()
			break
		}
		tag, value = p.splitTag(line)
		if tag == TagRelative {
			p.relative = strings.ToUpper(value) == "YES"
		} else if tag == TagBPM {
			parsed, err := parseFloat(value)
			if err != nil {
				return err
			}
			p.bpm = ultrastar.BPM(parsed * 4)
		} else if tag == TagEncoding {
			// Encoding tag is intentionally ignored
		} else if err := SetTag(song, tag, value); err != nil {
			return err
		}
	}
	return p.scanner.err()
}

func (p *Parser) splitTag(line string) (string, string) {
	var tag, value string
	index := strings.Index(line, ":")
	if index < 0 {
		tag, value = line[1:], ""
	} else {
		tag, value = line[1:index], line[index+1:]
	}
	return strings.ToUpper(strings.TrimSpace(tag)), strings.TrimSpace(value)
}

func (p *Parser) parseMusic(song *ultrastar.Song) error {
	var player int
	rel := [2]ultrastar.Beat{0, 0}
	song.MusicP1.SetBPM(p.bpm)

	if !p.scanner.scan() {
		return p.scanner.err()
	}
	line := p.scanner.text()
	if line[0] == 'P' {
		song.MusicP2 = ultrastar.NewMusicWithBPM(p.bpm)
	}
	p.scanner.unScan()
	music := [2]*ultrastar.Music{song.MusicP1, song.MusicP2}

LineLoop:
	for p.scanner.scan() {
		line = p.scanner.text()
		if line == "" {
			return ErrEmptyLine
		}
		switch line[0] {
		case uint8(ultrastar.NoteTypeRegular), uint8(ultrastar.NoteTypeGolden), uint8(ultrastar.NoteTypeFreestyle), uint8(ultrastar.NoteTypeRap), uint8(ultrastar.NoteTypeGoldenRap):
			note, err := ParseNote(line)
			if err != nil {
				return ErrInvalidNote
			}
			music[player].Notes = append(music[player].Notes, note)
		case '-':
			fields := strings.Fields(line[1:])
			if (!p.relative && len(fields) != 1) || (p.relative && len(fields) != 2) {
				return ErrInvalidLineBreak
			}
			beat, err := strconv.Atoi(fields[0])
			if err != nil {
				return ErrInvalidLineBreak
			}
			if p.relative {
				offset, err := strconv.Atoi(fields[1])
				if err != nil {
					return ErrInvalidLineBreak
				}
				rel[player] += ultrastar.Beat(offset)
			}
			music[player].LineBreaks = append(music[player].LineBreaks, rel[player]+ultrastar.Beat(beat))
		case 'P':
			if !song.IsDuet() {
				return ErrUnexpectedPNumber
			}
			p, err := strconv.Atoi(strings.TrimSpace(line[1:]))
			if err != nil || p < 1 || p > 2 {
				return ErrInvalidPNumber
			}
			player = p - 1
		case 'B':
			fields := strings.Fields(line[1:])
			if len(fields) != 2 {
				return ErrInvalidBPMChange
			}
			beat, err := strconv.Atoi(fields[0])
			if err != nil {
				return ErrInvalidBPMChange
			}
			bpm, err := parseFloat(fields[1])
			if err != nil {
				return ErrInvalidBPMChange
			}
			song.MusicP1.BPMs = append(song.MusicP1.BPMs, ultrastar.BPMChange{
				Start: rel[0] + ultrastar.Beat(beat),
				BPM:   ultrastar.BPM(bpm * 4),
			})
			if song.IsDuet() {
				// Even in duet mode BPM changes are always relative to P1
				song.MusicP2.BPMs = append(song.MusicP2.BPMs, ultrastar.BPMChange{
					Start: rel[0] + ultrastar.Beat(beat),
					BPM:   ultrastar.BPM(bpm * 4),
				})
			}
		case 'E':
			break LineLoop
		default:
			return ErrUnknownEvent
		}
	}
	song.MusicP1.Sort()
	if song.IsDuet() {
		song.MusicP2.Sort()
	}
	return p.scanner.err()
}

// parseFloat converts a string from an UltraStar txt to a float. This function
// implements some special parsing behavior to parse UltraStar floats,
// specifically supporting a comma as decimal separator.
func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(strings.Replace(s, ",", ".", 1), 64)
}
