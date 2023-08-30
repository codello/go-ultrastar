package txt

import (
	"bufio"
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

// ParseSong parses s into a song.
// This is a convenience method for [Reader.ReadSong].
func ParseSong(s string) (*ultrastar.Song, error) {
	return NewReader(strings.NewReader(s)).ReadSong()
}

// Reader implements the parser for the UltraStar TXT format.
type Reader struct {
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

	// Relative indicates whether the parser is in relative mode.
	// After parsing a song you can use this field to determine whether the song was originally in relative mode.
	Relative bool
	// Encoding is the encoding used to decode textual data.
	// During parsing this will be set to the appropriate header field of the song,
	// unless it has been set explicitly.
	Encoding string

	rd     io.Reader      //underlying reader
	s      *bufio.Scanner // s reads from rd
	rescan bool           // true indicates that the next scan operation should not advance the scanner
	line   string         // current line, set by scan
	lineNo int            // current line number, set by scan
	err    error          // last scanner error, set by scan
}

// NewReader creates a new Reader instance reading from rd.
// You can configure r before you start reading.
// After the first Read* call on the returned reader it is not guaranteed
// that configuration changes will be respected.
//
// The reader uses default settings that result in more strict parsing behavior
// compared to the UltraStar parser.
// Use [Reader.UseUltraStarDialect] to configure the reader to match UltraStar's parser more closely.
func NewReader(rd io.Reader) *Reader {
	r := &Reader{
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
	r.Reset(rd)
	return r
}

// UseUltraStarDialect configures r to match the behavior of the UltraStar TXT parser as closely as possible.
func (r *Reader) UseUltraStarDialect() {
	r.AllowBOM = true
	r.ApplyEncoding = true
	r.IgnoreEmptyLines = false
	r.IgnoreLeadingSpaces = false
	r.AllowRelative = true
	r.StrictLineBreaks = false
	r.EndTagRequired = false
	r.StrictEndTag = false
	r.AllowInternationalFloat = true
	r.IgnoreBPMChanges = true
}

// Reset configures r to read from r, just like NewReader(rd) would.
// r keeps its configuration, however r.Relative and r.Encoding are reset.
//
// Note that because Reader sometimes reads ahead, r.Reset(r.rd) may produce unexpected results.
func (r *Reader) Reset(rd io.Reader) {
	r.rd = rd
	r.s = nil
	r.rescan = false
	r.line = ""
	r.lineNo = 0
	r.err = nil

	r.Relative = false
	r.Encoding = ""
}

// setupScanner configures r.s.
// This must be called before any read operation is performed.
func (r *Reader) setupScanner() {
	if r.s == nil {
		if r.AllowBOM {
			r.rd = transform.NewReader(r.rd, unicode.BOMOverride(transform.Nop))
		}
		r.s = bufio.NewScanner(r.rd)
	}
}

// scan reads the next line of input.
// If r.rescan is true this operation does not advance the underlying scanner and r.line will not change.
// Otherwise, the underlying scanner is advanced and r.line and r.lineNo are updated accordingly.
func (r *Reader) scan() bool {
	if r.rescan {
		r.rescan = false
		return true
	}
	res := r.s.Scan()
	r.lineNo++

	if r.IgnoreEmptyLines {
		for res && strings.TrimSpace(r.s.Text()) == "" {
			res = r.s.Scan()
			r.lineNo++
		}
	}
	r.line = r.s.Text()
	r.err = r.s.Err()
	if r.IgnoreLeadingSpaces {
		r.line = strings.TrimLeft(r.line, " \t")
	}
	return res
}

// unscan sets a flag in r such that the next call to r.scan will not advance the underlying scanner.
// This effectively causes scan to read the same line again.
func (r *Reader) unscan() {
	if r.lineNo == 0 {
		panic("unscan called before scan.")
	}
	if r.rescan {
		panic("unscan called twice without scan.")
	}
	r.rescan = true
}

func (r *Reader) skipEmptyLines() error {
	if r.rescan && strings.TrimSpace(r.line) != "" {
		return nil
	}
	for r.scan() {
		if strings.TrimSpace(r.line) != "" {
			r.unscan()
			return nil
		}
	}
	return r.err
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
func (r *Reader) ReadSong() (*ultrastar.Song, error) {
	r.setupScanner()
	song, err := r.ReadTags()
	if err != nil {
		return song, ParseError{r.lineNo, err}
	}
	if err = r.skipEmptyLines(); err != nil {
		return song, ParseError{r.lineNo, r.err}
	}
	song.NotesP1, song.NotesP2, err = r.readNotes(true)
	if err != nil {
		return song, ParseError{r.lineNo, err}
	}
	if !r.ApplyEncoding {
		return song, nil
	}
	if err = r.applyEncoding(song); err != nil {
		return song, err
	}
	return song, nil
}

// ReadNotes parses an [ultrastar.Notes] from r.
// If the notes end with an end tag (a line starting with 'E') r may not be read until the end.
//
// If an error occurs this method may return a partial parse result up until the error occurred.
func (r *Reader) ReadNotes() (ultrastar.Notes, error) {
	notes, _, err := r.readNotes(false)
	if err != nil {
		return notes, ParseError{r.lineNo, err}
	}
	return notes, nil
}

// applyEncoding transforms all strings in s using the specified encoding name.
// The encoding name should identify a supported [charmap.Charmap].
// If the encoding is unknown or cannot be applied, the returned error will be non-nil.
func (r *Reader) applyEncoding(s *ultrastar.Song) error {
	var t transform.Transformer
	switch strings.ToLower(r.Encoding) {
	case "", "auto", "utf8", "utf-8":
		// This is the default
		return nil
	case "cp1250", "cp-1250", "windows1250", "windows-1250":
		t = charmap.Windows1250.NewDecoder()
	case "cp1252", "cp-1252", "windows1252", "windows-1252":
		t = charmap.Windows1252.NewDecoder()
	// FIXME: Do we want to support additional encodings?
	default:
		return ErrUnknownEncoding
	}

	return TransformSong(s, t)
}

// ReadTags reads a set of tags from the input and returns a song with the tags set.
// If an error occurs, it is returned.
func (r *Reader) ReadTags() (*ultrastar.Song, error) {
	r.setupScanner()
	song := &ultrastar.Song{}
	var tag, value string
	for r.scan() {
		if r.line == "" || r.line[0] != '#' {
			r.unscan()
			break
		}
		tag, value = splitTag(r.line)
		if tag == TagRelative {
			if !r.AllowRelative {
				return song, ErrRelativeNotAllowed
			}
			r.Relative = strings.ToUpper(value) == "YES"
		} else if tag == TagEncoding {
			if r.Encoding == "" {
				r.Encoding = value
			}
		} else if err := setTag(song, tag, value, r.AllowInternationalFloat); err != nil {
			return song, err
		}
	}
	return song, r.err
}

// splitTag is a helper method that splits a single tag line into key and value.
func splitTag(line string) (string, string) {
	var tag, value string
	index := strings.Index(line, ":")
	if index < 0 {
		tag, value = line[1:], ""
	} else {
		tag, value = line[1:index], line[index+1:]
	}
	return CanonicalTagName(strings.TrimSpace(tag)), strings.TrimSpace(value)
}

// readNotes parses the [ultrastar.Notes] of a song.
//
// allowDuet indicates whether scanning duets is allowed.
// If set to false a player change triggers an error.
func (r *Reader) readNotes(allowDuet bool) (ultrastar.Notes, ultrastar.Notes, error) {
	r.setupScanner()
	var (
		player int
		rel    [2]ultrastar.Beat
		notes  [2]ultrastar.Notes
	)

	if !r.scan() {
		return nil, nil, r.err
	}
	duet := r.line != "" && r.line[0] == 'P'
	r.unscan()

LineLoop:
	for r.scan() {
		if r.line == "" {
			return nil, nil, ErrEmptyLine
		}
		switch r.line[0] {
		case uint8(ultrastar.NoteTypeRegular), uint8(ultrastar.NoteTypeGolden), uint8(ultrastar.NoteTypeFreestyle), uint8(ultrastar.NoteTypeRap), uint8(ultrastar.NoteTypeGoldenRap):
			note, err := parseNoteRelative(r.line, r.Relative, r.StrictLineBreaks)
			if err != nil {
				return nil, nil, ErrInvalidNote
			}
			note.Start += rel[player]
			notes[player] = append(notes[player], note)
		case uint8(ultrastar.NoteTypeLineBreak):
			note, err := parseNoteRelative(r.line, r.Relative, r.StrictLineBreaks)
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
			p, err := strconv.Atoi(strings.TrimSpace(r.line[1:]))
			if err != nil || p < 1 || p > 2 {
				return nil, nil, ErrInvalidPNumber
			}
			player = p - 1
		case 'B':
			if !r.IgnoreBPMChanges {
				return nil, nil, ErrMultiBPM
			}
		case 'E':
			if r.StrictEndTag && strings.TrimSpace(r.line[1:]) != "" {
				return nil, nil, ErrInvalidEndTag
			}
			break LineLoop
		default:
			return nil, nil, fmt.Errorf("%c: %wr", r.line[0], ErrUnknownEvent)
		}
	}
	if r.err != nil {
		return nil, nil, r.err
	}
	if r.EndTagRequired && r.line[0] != 'E' {
		return nil, nil, ErrMissingEndTag
	}
	sort.Sort(notes[0])
	sort.Sort(notes[1])
	return notes[0], notes[1], nil
}
