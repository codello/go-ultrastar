package txt

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"

	"codello.dev/ultrastar"
)

// These are known errors that occur during parsing.
// Note that the parser wraps these in an [ParseError] value.
var (
	// ErrMultiBPM indicates that a song uses a BPM change marker which is not supported.
	ErrMultiBPM = errors.New("multi BPM songs are not supported")
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
	// ErrInvalidEndTag indicates that the end tag of a song was not correctly formatted.
	ErrInvalidEndTag = errors.New("invalid end tag")
	// ErrMissingEndTag indicates that no end tag was found.
	ErrMissingEndTag = errors.New("missing end tag")
	// ErrRelativeNotAllowed indicates that a song is in relative mode but the parser dialect forbids it.
	ErrRelativeNotAllowed = errors.New("RELATIVE tag not allowed")
	// ErrUnknownEvent indicates that the parser encountered a line starting with an unsupported symbol.
	ErrUnknownEvent = errors.New("invalid event")
	// ErrUnknownEncoding indicates that the value of the #ENCODING tag was not understood.
	ErrUnknownEncoding = errors.New("unknown encoding")
)

// A Dialect modifies the behavior of the parser.
// Using a Dialect you control how strict the parser will be when it comes to certain inaccuracies.
// This is analogous to the [Format] of the writer.
//
// Methods on Dialect values are safe for concurrent use by multiple goroutines
// as long as the dialect value remains unchanged.
type Dialect struct {
	// AllowBOM controls whether the parser should support songs that have an explicit Byte Order Mark.
	// If set to true the parser will understand and decode UTF-8 and UTF-16 BOMs.
	AllowBOM bool
	// ApplyEncoding controls whether the #ENCODING tag interpreted and applied to the song.
	// If it is not applied it will be treated as a custom tag.
	// If the encoding contains a value the parser does not understand it custom tag will be present as well.
	ApplyEncoding bool
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
	// IgnoreBPMChanges controls whether the parser silently ignores BPM change markers.
	IgnoreBPMChanges bool
}

// DialectDefault is the default dialect used for parsing UltraStar songs. The
// default dialect expects a more strict variant of songs.
var DialectDefault = Dialect{
	AllowBOM:                true,
	ApplyEncoding:           true,
	IgnoreEmptyLines:        true,
	IgnoreLeadingSpaces:     false,
	AllowRelative:           true,
	StrictLineBreaks:        true,
	EndTagRequired:          false,
	StrictEndTag:            true,
	AllowInternationalFloat: true,
	IgnoreBPMChanges:        false,
}

// DialectUltraStar is a parser dialect that very closely resembles the behavior
// of the TXT parser implementation of UltraStar Deluxe.
var DialectUltraStar = Dialect{
	AllowBOM:                true,
	ApplyEncoding:           true,
	IgnoreEmptyLines:        false,
	IgnoreLeadingSpaces:     false,
	AllowRelative:           true,
	StrictLineBreaks:        false,
	EndTagRequired:          false,
	StrictEndTag:            false,
	AllowInternationalFloat: true,
	IgnoreBPMChanges:        true,
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
// If an error occurs this method may return a partial result up until the error occurred.
// The concrete type of the error can be an instance of ParseError or TransformError
// indicating that the error occurred during parsing or decoding.
// It may also be an error value such as ErrUnknownEncoding.
func (d Dialect) ReadSong(r io.Reader) (*ultrastar.Song, error) {
	if d.AllowBOM {
		t := unicode.BOMOverride(transform.Nop)
		r = transform.NewReader(r, t)
	}
	p := newParser(r, d)

	song, err := p.parseTags()
	if err != nil {
		return song, ParseError{p.scanner.Line(), err}
	}
	if err := p.scanner.ScanEmptyLines(); err != nil {
		return song, ParseError{p.scanner.Line(), err}
	}
	song.NotesP1, song.NotesP2, err = p.parseNotes(true)
	if err != nil {
		return song, ParseError{p.scanner.Line(), err}
	}
	if err := d.applyEncoding(song, p.encoding); err != nil {
		return song, err
	}
	return song, nil
}

// ReadNotes parses an [ultrastar.Notes] from r.
// If the notes end with an end tag (a line starting with 'E') r may not be read until the end.
//
// If an error occurs this method may return a partial parse result up until the error occurred.
func (d Dialect) ReadNotes(r io.Reader) (ultrastar.Notes, error) {
	p := newParser(r, d)
	notes, _, err := p.parseNotes(false)
	if err != nil {
		return notes, ParseError{p.scanner.Line(), err}
	}
	return notes, nil
}

// applyEncoding transforms all strings in s using the specified encoding name.
// The encoding name should identify a supported charmap.Charmap.
// If the dialect does not enable encodings or if the encoding cannot be applied to the song without errors
// the song will have the #ENCODING custom tag set.
func (d Dialect) applyEncoding(s *ultrastar.Song, encoding string) error {
	if !d.ApplyEncoding {
		if encoding != "" {
			return d.SetTag(s, TagEncoding, encoding)
		}
		return nil
	}

	var t transform.Transformer
	switch strings.ToLower(encoding) {
	case "", "auto", "utf8", "utf-8":
		// This is the default
		return nil
	case "cp1250", "cp-1250", "windows1250", "windows-1250":
		t = charmap.Windows1250.NewDecoder()
	case "cp1252", "cp-1252", "windows1252", "windows-1252":
		t = charmap.Windows1252.NewDecoder()
	// FIXME: Do we want to support additional encodings?
	default:
		s.CustomTags[TagEncoding] = encoding
		return ErrUnknownEncoding
	}

	err := TransformSong(s, t)
	if err != nil {
		s.CustomTags[TagEncoding] = encoding
	}
	return err
}

// parser implements a parser for UltraStar songs.
// A parser is only valid for parsing a single [ultrastar.Song].
type parser struct {
	// A scanner for the underlying input
	scanner *scanner
	// The dialect used for parsing.
	dialect Dialect

	// relative indicates whether the parser is in relative mode.
	relative bool
	// encoding is the specified encoding that may be treated special.
	encoding string
}

// newParser creates a new parser reading from r using dialect d.
// The parser sets up its underlying scanner according to d.
// After constructing the parser changes to d may break parsing.
func newParser(r io.Reader, d Dialect) *parser {
	p := &parser{
		scanner: newScanner(r),
		dialect: d,
	}
	p.scanner.SkipEmptyLines = d.IgnoreEmptyLines
	p.scanner.TrimLeadingWhitespace = d.IgnoreLeadingSpaces
	return p
}

// parseTags reads the set of tags from the input and stores them into p.Song.
func (p *parser) parseTags() (*ultrastar.Song, error) {
	song := &ultrastar.Song{}
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
				return song, ErrRelativeNotAllowed
			}
			p.relative = strings.ToUpper(value) == "YES"
		} else if tag == TagEncoding {
			p.encoding = value
		} else if err := p.dialect.SetTag(song, tag, value); err != nil {
			return song, err
		}
	}
	return song, p.scanner.Err()
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

// parseNotes parses the [ultrastar.Notes] of a song.
//
// allowDuet indicates whether scanning duets is allowed.
// If set to false a player change triggers an error.
func (p *parser) parseNotes(allowDuet bool) (ultrastar.Notes, ultrastar.Notes, error) {
	var (
		player int
		rel    [2]ultrastar.Beat
		notes  [2]ultrastar.Notes
	)

	if !p.scanner.Scan() {
		return nil, nil, p.scanner.Err()
	}
	line := p.scanner.Text()
	duet := line[0] == 'P'
	p.scanner.UnScan()

LineLoop:
	for p.scanner.Scan() {
		line = p.scanner.Text()
		if line == "" {
			return nil, nil, ErrEmptyLine
		}
		switch line[0] {
		case uint8(ultrastar.NoteTypeRegular), uint8(ultrastar.NoteTypeGolden), uint8(ultrastar.NoteTypeFreestyle), uint8(ultrastar.NoteTypeRap), uint8(ultrastar.NoteTypeGoldenRap):
			note, err := p.dialect.ParseNoteRelative(line, p.relative)
			if err != nil {
				return nil, nil, ErrInvalidNote
			}
			note.Start += rel[player]
			notes[player] = append(notes[player], note)
		case uint8(ultrastar.NoteTypeLineBreak):
			note, err := p.dialect.ParseNoteRelative(line, p.relative)
			if err != nil {
				return nil, nil, ErrInvalidLineBreak
			}
			note.Start += rel[player]
			rel[player] += note.Duration
			note.Duration = 0
			notes[player] = append(notes[player], note)
		case 'P':
			if !allowDuet || !duet {
				return nil, nil, ErrUnexpectedPNumber
			}
			p, err := strconv.Atoi(strings.TrimSpace(line[1:]))
			if err != nil || p < 1 || p > 2 {
				return nil, nil, ErrInvalidPNumber
			}
			player = p - 1
		case 'B':
			if !p.dialect.IgnoreBPMChanges {
				return nil, nil, ErrMultiBPM
			}
		case 'E':
			if p.dialect.StrictEndTag && strings.TrimSpace(line[1:]) != "" {
				return nil, nil, ErrInvalidEndTag
			}
			break LineLoop
		default:
			return nil, nil, ErrUnknownEvent
		}
	}
	if p.scanner.Err() != nil {
		return nil, nil, p.scanner.Err()
	}
	if p.dialect.EndTagRequired && line[0] != 'E' {
		return nil, nil, ErrMissingEndTag
	}
	sort.Sort(notes[0])
	sort.Sort(notes[1])
	return notes[0], notes[1], nil
}

// ParseError is an error type that may be returned by the parsing methods.
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
