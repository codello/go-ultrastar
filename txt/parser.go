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
	DialectDefault = &Dialect{
		SkipEmptyLines:          true,
		TrimLeadingSpace:        false,
		AllowRelative:           true,
		StrictLineBreaks:        true,
		AllowInternationalFloat: true,
	}
	DialectUltraStar = &Dialect{
		SkipEmptyLines:          false,
		TrimLeadingSpace:        false,
		AllowRelative:           true,
		StrictLineBreaks:        false,
		AllowInternationalFloat: true,
	}

	// ErrEmptyLine indicates that an empty line was encountered but not expected.
	ErrEmptyLine          = errors.New("unexpected empty line")
	ErrInvalidPNumber     = errors.New("invalid P-number")
	ErrUnexpectedPNumber  = errors.New("unexpected P number")
	ErrInvalidNote        = errors.New("invalid note")
	ErrInvalidLineBreak   = errors.New("invalid line break")
	ErrInvalidBPMChange   = errors.New("invalid BPM change")
	ErrRelativeNotAllowed = errors.New("RELATIVE tag not allowed")
	ErrUnknownEvent       = errors.New("invalid event")
)

func ParseSong(s string) (*ultrastar.Song, error) {
	return ReadSong(strings.NewReader(s))
}

func ReadSong(r io.Reader) (*ultrastar.Song, error) {
	return DialectDefault.ReadSong(r)
}

// TODO: Implement note parsing dialect
type Dialect struct {
	SkipEmptyLines          bool
	TrimLeadingSpace        bool
	AllowRelative           bool
	StrictLineBreaks        bool
	AllowInternationalFloat bool
}

func (d *Dialect) ReadSong(r io.Reader) (*ultrastar.Song, error) {
	// TODO: Doc: May return partial result
	p := newParser(r, d)

	if err := p.parseTags(); err != nil {
		return p.Song, &ParseError{p.scanner.Line(), err}
	}
	if err := p.scanner.ScanEmptyLines(); err != nil {
		return p.Song, &ParseError{p.scanner.Line(), err}
	}
	if err := p.parseMusic(true); err != nil {
		return p.Song, &ParseError{p.scanner.Line(), err}
	}
	return p.Song, nil
}

func (d *Dialect) ReadMusic(r io.Reader) (*ultrastar.Music, error) {
	p := newParser(r, d)
	if err := p.parseMusic(false); err != nil {
		return p.Song.MusicP1, &ParseError{p.scanner.Line(), err}
	}
	return p.Song.MusicP1, nil
}

type parser struct {
	scanner *scanner
	dialect *Dialect

	relative bool
	bpm      ultrastar.BPM

	Song *ultrastar.Song
}

func newParser(r io.Reader, d *Dialect) *parser {
	p := &parser{
		scanner: newScanner(r),
		dialect: d,
		Song:    ultrastar.NewSong(),
	}
	p.scanner.SkipEmptyLines = d.SkipEmptyLines
	p.scanner.TrimLeadingWhitespace = d.TrimLeadingSpace
	return p
}

func (p *parser) parseTags() error {
	var line, tag, value string
	for p.scanner.Scan() {
		line = p.scanner.Text()
		if line == "" || line[0] != '#' {
			p.scanner.UnScan()
			break
		}
		tag, value = p.splitTag(line)
		if tag == TagRelative {
			if !p.dialect.AllowRelative {
				return ErrRelativeNotAllowed
			}
			p.relative = strings.ToUpper(value) == "YES"
		} else if tag == TagBPM {
			parsed, err := p.dialect.parseFloat(value)
			if err != nil {
				return err
			}
			p.bpm = ultrastar.BPM(parsed * 4)
		} else if tag == TagEncoding {
			// Encoding tag is intentionally ignored
		} else if err := p.dialect.SetTag(p.Song, tag, value); err != nil {
			return err
		}
	}
	return p.scanner.Err()
}

func (p *parser) splitTag(line string) (string, string) {
	var tag, value string
	index := strings.Index(line, ":")
	if index < 0 {
		tag, value = line[1:], ""
	} else {
		tag, value = line[1:index], line[index+1:]
	}
	return CanonicalTagName(strings.TrimSpace(tag)), strings.TrimSpace(value)
}

func (p *parser) parseMusic(allowDuet bool) error {
	player := 0
	rel := [2]ultrastar.Beat{0, 0}
	p.Song.MusicP1.SetBPM(p.bpm)

	if !p.scanner.Scan() {
		return p.scanner.Err()
	}
	line := p.scanner.Text()
	if line[0] == 'P' {
		p.Song.MusicP2 = ultrastar.NewMusicWithBPM(p.bpm)
	}
	p.scanner.UnScan()
	music := [2]*ultrastar.Music{p.Song.MusicP1, p.Song.MusicP2}

LineLoop:
	for p.scanner.Scan() {
		line = p.scanner.Text()
		if line == "" {
			return ErrEmptyLine
		}
		switch line[0] {
		case uint8(ultrastar.NoteTypeRegular), uint8(ultrastar.NoteTypeGolden), uint8(ultrastar.NoteTypeFreestyle), uint8(ultrastar.NoteTypeRap), uint8(ultrastar.NoteTypeGoldenRap):
			note, err := p.dialect.ParseNoteRelative(line, p.relative)
			if err != nil {
				return ErrInvalidNote
			}
			note.Start += rel[player]
			music[player].Notes = append(music[player].Notes, note)
		case uint8(ultrastar.NoteTypeLineBreak):
			note, err := p.dialect.ParseNoteRelative(line, p.relative)
			if err != nil {
				return ErrInvalidLineBreak
			}
			note.Start += rel[player]
			rel[player] += note.Duration
			note.Duration = 0
			music[player].Notes = append(music[player].Notes, note)
		case 'P':
			if !allowDuet || !p.Song.IsDuet() {
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
			bpm, err := p.dialect.parseFloat(fields[1])
			if err != nil {
				return ErrInvalidBPMChange
			}
			p.Song.MusicP1.BPMs = append(p.Song.MusicP1.BPMs, ultrastar.BPMChange{
				Start: rel[0] + ultrastar.Beat(beat),
				BPM:   ultrastar.BPM(bpm * 4),
			})
			if p.Song.IsDuet() {
				// Even in duet mode BPM changes are always relative to P1
				p.Song.MusicP2.BPMs = append(p.Song.MusicP2.BPMs, ultrastar.BPMChange{
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
	p.Song.MusicP1.Sort()
	if p.Song.IsDuet() {
		p.Song.MusicP2.Sort()
	}
	return p.scanner.Err()
}

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
