package txt

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/Karaoke-Manager/go-ultrastar"
)

// These are known errors that occur during parsing.
// Note that the parser wraps these in an [ParseError] value.
var (
	// ErrEmptyLine indicates that an empty line was encountered but not expected.
	ErrEmptyLine = errors.New("unexpected empty line")
	// ErrInvalidPNumber indicates that a player change was not correctly formatted.
	ErrInvalidPNumber = errors.New("invalid P-number")
	// ErrUnexpectedPNumber indicates that a player change was found in a non-duet song.
	ErrUnexpectedPNumber = errors.New("unexpected P number")
	// ErrInvalidNote indicates that a note could not be parsed.
	ErrInvalidNote = errors.New("invalid note")
	// ErrInvalidLineBreak indicates that a line break could not be parsed.
	ErrInvalidLineBreak = errors.New("invalid line break")
	// ErrInvalidBPMChange indicates that a BPM change could not be parsed.
	ErrInvalidBPMChange = errors.New("invalid BPM change")
	// ErrInvalidEndTag indicates that the end tag of a song was not correctly formatted.
	ErrInvalidEndTag = errors.New("invalid end tag")
	// ErrMissingEndTag indicates that no end tag was found.
	ErrMissingEndTag = errors.New("missing end tag")
	// ErrRelativeNotAllowed indicates that a song is in relative mode but the parser dialect forbids it.
	ErrRelativeNotAllowed = errors.New("RELATIVE tag not allowed")
	// ErrUnknownEvent indicates that the parser encountered a line starting with an unsupported symbol.
	ErrUnknownEvent = errors.New("invalid event")
)

// A Dialect modifies the behavior of the parser.
// Using a Dialect you control how strict the parser will be when it comes to certain inaccuracies.
// This is analogous to the [Format] of the writer.
//
// Methods on Dialect values are safe for concurrent use by multiple goroutines
// as long as the dialect value remains unchanged.
type Dialect struct {
	// IgnoreEmptyLines specifies whether empty lines are allowed in songs.
	IgnoreEmptyLines bool
	// IgnoreLeadingSpaces controls whether leading spaces are ignored in songs.
	// This only applies to the beginning of lines and not e.g. to note texts.
	IgnoreLeadingSpaces bool
	// AllowRelative controls whether the parser allows parsing of songs in relative mode.
	AllowRelative bool
	// StrictLineBreaks controls whether line breaks can have additional text on their line.
	// If set to true the parser will return an error if a line break has additional text.
	StrictLineBreaks bool
	// EndTagRequired controls whether the final 'E' is required.
	EndTagRequired bool
	// StrictEndTag controls whether any line starting with 'E' counts as an end tag.
	// If set to true only a single 'E' may be on the ending line.
	StrictEndTag bool
	// AllowInternationalFloat controls whether floats can use a comma as the decimal separator.
	AllowInternationalFloat bool
}

// DialectDefault is the default dialect used for parsing UltraStar songs. The
// default dialect expects a more strict variant of songs.
var DialectDefault = &Dialect{
	IgnoreEmptyLines:        true,
	IgnoreLeadingSpaces:     false,
	AllowRelative:           true,
	StrictLineBreaks:        true,
	EndTagRequired:          false,
	StrictEndTag:            true,
	AllowInternationalFloat: true,
}

// DialectUltraStar is a parser dialect that very closely resembles the behavior
// of the TXT parser implementation of UltraStar Deluxe.
var DialectUltraStar = &Dialect{
	IgnoreEmptyLines:        false,
	IgnoreLeadingSpaces:     false,
	AllowRelative:           true,
	StrictLineBreaks:        false,
	EndTagRequired:          false,
	StrictEndTag:            false,
	AllowInternationalFloat: true,
}

// ParseSong parses s into a song.
// This is a convenience method for [Dialect.ReadSong].
func ParseSong(s string) (*ultrastar.Song, error) {
	return ReadSong(strings.NewReader(s))
}

// ReadSong parses a song from r.
// This is a convenience method for [Dialect.ReadSong].
//
// Note that r is not necessarily read completely.
func ReadSong(r io.Reader) (*ultrastar.Song, error) {
	return DialectDefault.ReadSong(r)
}

// ReadSong parses an [ultrastar.Song] from r.
// If the song ends with an end tag (a line starting with 'E') r may not be read until the end.
//
// The song returned by this method will always be in absolute time.
// If the source file uses relative mode the times will be converted to absolute times.
//
// If an error occurs this method may return a partial parse result up until the error occurred.
func (d *Dialect) ReadSong(r io.Reader) (*ultrastar.Song, error) {
	p := newParser(r, d)

	if err := p.parseTags(); err != nil {
		return p.Song, ParseError{p.scanner.Line(), err}
	}
	if err := p.scanner.ScanEmptyLines(); err != nil {
		return p.Song, ParseError{p.scanner.Line(), err}
	}
	if err := p.parseMusic(true); err != nil {
		return p.Song, ParseError{p.scanner.Line(), err}
	}
	return p.Song, nil
}

// ReadMusic parses an [ultrastar.Music] from r.
// If the music ends with an end tag (a line starting with 'E') r may not be read until the end.
//
// If an error occurs this method may return a partial parse result up until the error occurred.
func (d *Dialect) ReadMusic(r io.Reader) (*ultrastar.Music, error) {
	p := newParser(r, d)
	if err := p.parseMusic(false); err != nil {
		return p.Song.MusicP1, ParseError{p.scanner.Line(), err}
	}
	return p.Song.MusicP1, nil
}

// parser implements a parser for UltraStar songs.
// A parser is only valid for parsing a single [ultrastar.Song].
type parser struct {
	// A scanner for the underlying input
	scanner *scanner
	// The dialect used for parsing.
	dialect *Dialect

	// relative indicates whether the parser is in relative mode.
	relative bool
	// bpm stores the BPM value parsed from the song header.
	bpm ultrastar.BPM

	// Song is the song constructed by the parser.
	Song *ultrastar.Song
}

// newParser creates a new parser reading from r using dialect d.
// The parser sets up its underlying scanner according to d.
// After constructing the parser changes to d may break parsing.
func newParser(r io.Reader, d *Dialect) *parser {
	p := &parser{
		scanner: newScanner(r),
		dialect: d,
		Song:    ultrastar.NewSong(),
	}
	p.scanner.SkipEmptyLines = d.IgnoreEmptyLines
	p.scanner.TrimLeadingWhitespace = d.IgnoreLeadingSpaces
	return p
}

// parseTags reads the set of tags from the input and stores them into p.Song.
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

// splitTag is a helper method that splits a single tag line into key and value.
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

// parseMusic parses the [ultrastar.Music] of a song.
//
// allowDuet indicates whether scanning duets is allowed.
// If set to false a player change triggers an error.
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
			if p.dialect.StrictEndTag && strings.TrimSpace(line[1:]) != "" {
				return ErrInvalidEndTag
			}
			break LineLoop
		default:
			return ErrUnknownEvent
		}
	}
	if p.scanner.Err() != nil {
		return p.scanner.Err()
	}
	if p.dialect.EndTagRequired && line[0] != 'E' {
		return ErrMissingEndTag
	}
	p.Song.MusicP1.Sort()
	if p.Song.IsDuet() {
		p.Song.MusicP2.Sort()
	}
	return nil
}

// ParseError is the error type returned by the parsing methods.
// It wraps an underlying error and also provides a line number on which the error occurred.
type ParseError struct {
	// line is the line number that caused the error.
	line int
	// err is the underlying error.
	err error
}

// Error returns the error string.
func (e ParseError) Error() string {
	return fmt.Sprintf("parse error at line %d: %v", e.Line(), e.err)
}

// Line returns the line number that caused the error.
func (e ParseError) Line() int {
	return e.line
}

// Unwrap returns the underlying error.
func (e ParseError) Unwrap() error {
	return e.err
}
