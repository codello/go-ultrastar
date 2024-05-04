package ultrastar

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	xunicode "golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// Encodings is a map of known encodings that are recognized by the [Reader].
// You can add or remove encodings to configure the known encodings.
//
// Read and write access is not synchronized between multiple goroutines.
var Encodings = map[string]encoding.Encoding{
	"":             nil,
	"auto":         nil,
	"utf8":         nil,
	"utf-8":        nil,
	"cp1250":       charmap.Windows1250,
	"cp-1250":      charmap.Windows1250,
	"windows1250":  charmap.Windows1250,
	"windows-1250": charmap.Windows1250,
	"cp1252":       charmap.Windows1252,
	"cp-1252":      charmap.Windows1252,
	"windows1252":  charmap.Windows1252,
	"windows-1252": charmap.Windows1252,
}

type SyntaxError struct {
	Line int
	err  error
}

// Error returns the error string.
func (e *SyntaxError) Error() string {
	return fmt.Sprintf("syntax error at line %d: %v", e.Line, e.err)
}

// Unwrap returns the underlying error.
func (e *SyntaxError) Unwrap() error {
	return e.err
}

type HeaderError struct {
	Key string
	Err error
}

func (e *HeaderError) Error() string {
	return fmt.Sprintf("invalid value for header #%s: %s", e.Key, e.Err)
}

func (e *HeaderError) Unwrap() error {
	return e.Err
}

type EncodeError struct {
	Key   string
	Value string
	Text  string
}

// ParseSong parses s into a song.
// This is a convenience method for [Reader.ReadSong].
func ParseSong(s string) (*Song, error) {
	r, err := NewReader(strings.NewReader(s))
	if err != nil {
		return nil, err
	}
	return r.ReadSong()
}

// Reader implements the parser for the UltraStar TXT format.
type Reader struct {
	// Header contains the raw header values read by a Reader.
	// A Header is only valid until [Reader.Reset] is called.
	// If you need to access the headers afterward, you must make a copy first.
	Header Header

	Version  Version
	Relative bool
	Encoding encoding.Encoding

	s      *bufio.Scanner // s reads from rd
	rescan bool
	line   int // current line number, set by scan
	voice  int
	rel    []Beat
}

// NewReader creates a new Reader instance and calls [Reader.Reset].
// Any error during the reset is returned by this function.
// Regardless of error, a valid Reader is returned.
//
// See [Reader.Reset] for details on possible errors.
func NewReader(rd io.Reader) (*Reader, error) {
	r := &Reader{}
	return r, r.Reset(rd)
}

// Reset resets the internal state of r and reads a file header from r.
// After this method returns, r.Header is valid.
// This method sets r.Version, r.Relative and r.Encoding according to the file header.
//
// There are multiple kinds of errors that can occur during the reset process.
// Reset will try its best to process the complete header.
// The returned error can be unwrapped to access the individual error values:
//
// TODO: Docs???
//   - A [SyntaxError] indicates a header line that was not formatted correctly.
//   - ErrUnknownEncoding
//   - An [EncodeError] indicates an error when applying the encoding header
func (r *Reader) Reset(rd io.Reader) error {
	r.s = bufio.NewScanner(transform.NewReader(rd, xunicode.BOMOverride(transform.Nop)))
	r.s.Split(scanLines)
	r.rescan = false
	r.line = 0
	r.voice = P1
	r.rel = make([]Beat, 9)

	clear(r.Header)
	r.Relative = false
	r.Version = Version030
	r.Encoding = nil

	return r.readHeader()
}

// scanLines is a split function for a Scanner that returns each line of text,
// stripped of any trailing end-of-line marker.
// The returned line may be empty. The end-of-line marker is a single newline, a single carriage return, or a
// carriage return followed by a newline.
// The last non-empty line of input will be returned even if it has no newline.
func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexAny(data, "\r\n"); i >= 0 {
		if data[i] == '\n' {
			// We have a line terminated by a single newline
			return i + 1, data[0:i], nil
		}
		if !atEOF && len(data) == i+1 {
			// We have a carriage return at the end of the buffer
			// Request more data to see if a newline follows
			return 0, nil, nil
		}
		// We have a line terminated by a carriage return
		advance = i + 1
		if len(data) > i+1 && data[i+1] == '\n' {
			// The carriage return is followed by a newline
			advance++
		}
		return advance, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

// Line returns the number of lines that have already been processed by r.
func (r *Reader) Line() int {
	return r.line
}

var ErrUnknownEncoding = errors.New("unknown encoding")

func (r *Reader) readHeader() error {
	for ok := true; ok; {
		var (
			err  error
			sErr *SyntaxError
		)
		if ok, err = r.readHeaderLine(); errors.As(err, &sErr) {
			// errs = append(errs, err)
			// FIXME: Is it ok to ignore syntax errors here?
		} else if err != nil {
			// Errors from the underlying reader are non-recoverable
			return err
		}
	}
	if v, err := getUniqueHeaderAs(r.Header.Values(HeaderVersion), true, ParseVersion); errors.Is(err, ErrNoValue) {
		r.Version = Version030
	} else if err != nil {
		return &HeaderError{HeaderVersion, err}
	} else {
		r.Version = v
	}
	if r.Version.LessThan(Version100) {
		var relErr error
		r.Relative, relErr = getUniqueHeaderAs(r.Header.Values(HeaderRelative), false, func(v string) (bool, error) {
			return strings.EqualFold(v, "yes"), nil
		})
		enc, encErr := getUniqueHeaderAs(r.Header.Values(HeaderEncoding), false, func(v string) (encoding.Encoding, error) {
			if enc, ok := Encodings[strings.ToLower(v)]; ok {
				return enc, nil
			} else {
				return nil, ErrUnknownEncoding
			}
		})
		if encErr == nil {
			encErr = r.UseEncoding(enc)
		}
		if encErr != nil {
			return &HeaderError{HeaderEncoding, encErr}
		}
		if relErr != nil {
			return &HeaderError{HeaderRelative, relErr}
		}
	}
	return nil
}

func (r *Reader) readHeaderLine() (bool, error) {
	if !r.scan() {
		return false, r.s.Err()
	}
	line := r.s.Bytes()
	if len(bytes.TrimSpace(line)) == 0 {
		return true, nil
	}
	if line[0] != '#' {
		r.unscan()
		return false, nil
	}
	// FIXME: Are these actually errors? Or is there an interpretation? Or should header errors be ignored by spec?
	index := bytes.IndexByte(line, ':')
	if index < 0 {
		return true, &SyntaxError{r.line, errors.New("colon expected, got end of line")}
	}
	key := bytes.TrimSpace(line[1:index])
	if len(key) == 0 {
		return true, &SyntaxError{r.line, errors.New("empty header key")}
	}
	r.Header.Add(
		string(key),
		string(bytes.TrimSpace(line[index+1:])),
	)
	return true, nil
}

func (r *Reader) ReadNote() (Note, int, error) {
	var line []byte

	// There can be multiple voice changes without a single note.
	// Only the last voice change is significant.
	for {
		if !r.scan() {
			if r.s.Err() == nil {
				return Note{}, -1, io.EOF
			}
			return Note{}, r.voice, r.s.Err()
		}
		line = r.s.Bytes()
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 {
			// ignore empty lines
			continue
		}
		switch line[0] {
		case 'E':
			if len(trimmed) > 1 {
				return Note{}, -1, &SyntaxError{r.line, errors.New("invalid end tag")}
			}
			return Note{}, -1, io.EOF
		case 'P':
			voice, err := strconv.Atoi(string(trimmed[1:]))
			if err != nil || voice <= 0 {
				return Note{}, -1, &SyntaxError{r.line, errors.New("invalid voice change")}
			}
			r.voice = voice - 1
			continue
		}
		break
	}

	note, err := r.parseNote(line)
	if err != nil {
		return note, r.voice, &SyntaxError{r.line, errors.New("invalid note")}
	}
	return note, r.voice, nil
}

func (r *Reader) parseNote(line []byte) (Note, error) {
	note := Note{}
	if len(line) == 0 {
		return note, errors.New("invalid note type")
	}
	note.Type = NoteType(line[0])
	line = line[1:]
	//if note.Type.IsStandard() {
	// FIXME: If custom note types are forbidden, raise an error here
	// return note, &SyntaxError{r.line, errors.New("invalid note type")}
	//}
	if note.Type == NoteTypeEndOfPhrase {
		note.Text = "\n"
	}

	// note start
	value, line := nextField(line)
	start, err := strconv.Atoi(value)
	if err != nil {
		return note, fmt.Errorf("invalid note start: %w", err)
	}
	note.Start = Beat(start) + r.rel[r.voice]

	// end-of-phrase
	if note.Type == NoteTypeEndOfPhrase && !r.Relative {
		if len(bytes.TrimSpace(line)) > 0 {
			return note, fmt.Errorf("invalid line break: extra text")
		}
		return note, nil
	}

	// note duration or relative offset
	value, line = nextField(line)
	duration, err := strconv.Atoi(value)
	if note.Type == NoteTypeEndOfPhrase {
		if err != nil {
			return note, fmt.Errorf("invalid line break: invalid relative offset: %w", err)
		}
		if len(bytes.TrimSpace(line)) > 0 {
			return note, fmt.Errorf("invalid line break: extra text")
		}
		r.rel[r.voice] += Beat(duration)
		return note, nil
	}
	if err != nil {
		return note, fmt.Errorf("invalid note duration: %w", err)
	}
	note.Duration = Beat(duration)

	// note pitch
	value, line = nextField(line)
	pitch, err := strconv.Atoi(value)
	if err != nil {
		return note, fmt.Errorf("invalid note pitch: %w", err)
	}
	note.Pitch = Pitch(pitch)

	// note text
	rne, s := utf8.DecodeRune(line)
	line = line[s:]
	if !unicode.Is(unicode.White_Space, rne) {
		return note, errors.New("missing whitespace after note pitch")
	}
	if len(line) == 0 {
		return note, errors.New("empty note text")
	}
	if r.Encoding != nil {
		line, err = r.Encoding.NewDecoder().Bytes(line)
		if err != nil {
			return note, err
		}
	}
	note.Text = string(line)
	return note, nil
}

func nextField(line []byte) (string, []byte) {
	line = bytes.TrimLeftFunc(line, func(r rune) bool {
		return unicode.Is(unicode.White_Space, r)
	})
	i := bytes.IndexFunc(line, func(r rune) bool {
		return !unicode.Is(unicode.White_Space, r)
	})
	if i < 0 {
		i = len(line)
	}
	return string(line[:i]), line[i:]
}

// ReadSong parses a [Song] from r.
// If the song ends with an end tag (a line starting with 'E') r may not be read until the end.
//
// The song returned by this method will always be in absolute time.
// If the source file uses relative mode the times will be converted to absolute times.
//
// If an error occurs this method may return a partial result up until the error occurred.
// The concrete type of the error can be an instance of SyntaxError or TransformError
// indicating that the error occurred during parsing or decoding.
// It may also be an error value such as ErrUnknownEncoding.
func (r *Reader) ReadSong() (*Song, error) {
	song := &Song{}
	if err := r.readSongHeader(song); err != nil {
		return nil, err
	}

	song.Voices = make([]*Voice, 9)
	for i := range song.Voices {
		name := r.Header.Get("P" + strconv.Itoa(i+1))
		if name == "" && r.Version.Compare(Version100) >= 0 {
			name = r.Header.Get("DUETSINGERP" + strconv.Itoa(i+1))
		}
		song.Voices[i] = &Voice{Name: name}
	}
	for {
		if note, voice, err := r.ReadNote(); err == nil {
			song.Voices[voice].AppendNotes(note)
		} else {
			if errors.Is(err, io.EOF) {
				err = nil
			}
			slices.DeleteFunc(song.Voices, func(voice *Voice) bool {
				return voice.Name == "" && voice.IsEmpty()
			})
			clear(song.Voices[len(song.Voices):cap(song.Voices)])
			return song, err
		}
	}
}

func (r *Reader) readSongHeader(s *Song) error {
	// Other headers depend on GAP and BPM so these need to be processed first
	if err := r.setSongHeader(s, HeaderBPM, r.Header.Values(HeaderBPM)); err != nil {
		return &HeaderError{HeaderBPM, err}
	}
	if err := r.setSongHeader(s, HeaderGap, r.Header.Values(HeaderGap)); err != nil {
		return &HeaderError{HeaderGap, err}
	}
	for k, vs := range r.Header {
		if k == HeaderBPM || k == HeaderGap {
			continue
		}
		if err := r.setSongHeader(s, k, vs); err != nil {
			return &HeaderError{k, err}
		}
	}
	return nil
}

func (r *Reader) setSongHeader(s *Song, key string, values []string) (err error) {
	switch key {
	case HeaderVersion, "P1", "P2", "P3", "P4", "P5", "P6", "P7", "P8", "P9":
		// These headers are recognized by the Reader and are not passed to s.
	case HeaderEncoding, HeaderRelative,
		"DUETSINGERP1", "DUETSINGERP2", "DUETSINGERP3", "DUETSINGERP4", "DUETSINGERP5", "DUETSINGERP6", "DUETSINGERP7", "DUETSINGERP8", "DUETSINGERP9":
		// These headers are recognized by the Reader and are not passed to s.
		if r.Version.Compare(Version100) >= 0 || r.Version.LessThan(Version010) {
			s.ExtraHeaders.SetValues(key, values)
		}
	case HeaderBPM:
		s.BPM, err = getUniqueHeaderAs(values, true, func(v string) (BPM, error) {
			// TODO: Is comma allowed in v2?
			f, err := parseFloat(v, true)
			b := BPM(f) * 4
			if !b.IsValid() {
				return b, fmt.Errorf("invalid BPM value: %f", b)
			}
			return b, err
		})
	case HeaderGap:
		if r.Version.LessThan(Version200) {
			s.Gap, err = getUniqueHeaderAs(values, false, func(v string) (time.Duration, error) {
				f, err := parseFloat(v, true)
				return time.Duration(f * float64(time.Millisecond)), err
			})
		} else if r.Version.GreaterThan(Version010) {
			s.Gap, err = getUniqueHeaderAs(values, false, parseDurationMilliseconds)
		} else {
			s.ExtraHeaders.SetValues(key, values)
		}
	case HeaderMP3:
		if r.Version.LessThan(Version200) {
			if r.Version.LessThan(Version110) || len(s.Audio) == 0 {
				s.Audio = parseMultiValuedHeader(values)
			}
		} else {
			s.ExtraHeaders.SetValues(key, values)
		}
	case HeaderAudio:
		if r.Version.Compare(Version110) >= 0 {
			s.Audio = parseMultiValuedHeader(values)
		} else {
			s.ExtraHeaders.SetValues(key, values)
		}
	case HeaderVocals:
		if r.Version.Compare(Version110) >= 0 {
			s.Vocals = parseMultiValuedHeader(values)
		} else {
			s.ExtraHeaders.SetValues(key, values)
		}
	case HeaderInstrumental:
		if r.Version.Compare(Version110) >= 0 {
			s.Instrumental = parseMultiValuedHeader(values)
		} else {
			s.ExtraHeaders.SetValues(key, values)
		}
	case HeaderVideo:
		if r.Version.Compare(Version020) >= 0 {
			s.Video = parseMultiValuedHeader(values)
		} else {
			s.ExtraHeaders.SetValues(key, values)
		}
	case HeaderCover:
		if r.Version.Compare(Version020) >= 0 {
			s.Cover = parseMultiValuedHeader(values)
		} else {
			s.ExtraHeaders.SetValues(key, values)
		}
	case HeaderBackground:
		if r.Version.Compare(Version020) >= 0 {
			s.Background = parseMultiValuedHeader(values)
		} else {
			s.ExtraHeaders.SetValues(key, values)
		}
	case HeaderVideoGap:
		if r.Version.Compare(Version200) >= 0 {
			s.VideoGap, err = getUniqueHeaderAs(values, false, parseDurationMilliseconds)
		} else if r.Version.Compare(Version020) >= 0 {
			s.VideoGap, err = getUniqueHeaderAs(values, false, parseDurationSeconds)
		} else {
			s.ExtraHeaders.SetValues(key, values)
		}
	case HeaderStart:
		if r.Version.Compare(Version200) >= 0 {
			s.Start, err = getUniqueHeaderAs(values, false, parseDurationMilliseconds)
		} else if r.Version.Compare(Version020) >= 0 {
			s.Start, err = getUniqueHeaderAs(values, false, parseDurationSeconds)
		} else {
			s.ExtraHeaders.SetValues(key, values)
		}
	case HeaderEnd:
		if r.Version.Compare(Version020) >= 0 {
			s.End, err = getUniqueHeaderAs(values, false, parseDurationMilliseconds)
		} else {
			s.ExtraHeaders.SetValues(key, values)
		}
	case HeaderCalcMedley:
		if r.Version.Compare(Version020) >= 0 {
			s.NoAutoMedley, err = getUniqueHeaderAs(values, false, func(v string) (bool, error) {
				return strings.EqualFold(v, "off"), nil
			})
		} else {
			s.ExtraHeaders.SetValues(key, values)
		}
	case HeaderMedleyStart:
		// FIXME: Plus GAP or not?
		if r.Version.LessThan(Version200) {
			s.ExtraHeaders.SetValues(key, values)
		} else {
			s.MedleyStart, err = getUniqueHeaderAs(values, false, parseDurationMilliseconds)
		}
	case HeaderMedleyEnd:
		// FIXME: Plus GAP or not?
		if r.Version.LessThan(Version200) {
			s.ExtraHeaders.SetValues(key, values)
		} else {
			s.MedleyEnd, err = getUniqueHeaderAs(values, false, parseDurationMilliseconds)
		}
	case HeaderMedleyStartBeat:
		if r.Version.LessThan(Version200) && r.Version.Compare(Version020) >= 0 {
			if st, pErr := getUniqueHeaderAs(values, true, strconv.Atoi); pErr != nil && !errors.Is(pErr, ErrNoValue) {
				err = pErr
			} else if err == nil {
				s.MedleyStart = s.BPM.Duration(Beat(st)) + s.Gap
			}
		} else {
			s.ExtraHeaders.SetValues(key, values)
		}
	case HeaderMedleyEndBeat:
		if r.Version.LessThan(Version200) && r.Version.Compare(Version020) >= 0 {
			if e, pErr := getUniqueHeaderAs(values, true, strconv.Atoi); pErr != nil && !errors.Is(pErr, ErrNoValue) {
				err = pErr
			} else if err == nil {
				s.MedleyEnd = s.BPM.Duration(Beat(e)) + s.Gap
			}
		} else {
			s.ExtraHeaders.SetValues(key, values)
		}
	case HeaderTitle:
		s.Title, err = getUniqueHeaderAs(values, false, func(v string) (string, error) {
			return v, nil
		})
	case HeaderArtist:
		// TODO: Handle multiple artists (@spec)
		s.Artist, err = getUniqueHeaderAs(values, false, func(v string) (string, error) {
			return v, nil
		})
	case HeaderGenre:
		// FIXME: How to handle multi-values for earlier versions?
		if r.Version.Compare(Version020) >= 0 {
			s.Genres = parseMultiValuedHeader(values)
		}
	case HeaderEdition:
		if r.Version.Compare(Version020) >= 0 {
			s.Editions = parseMultiValuedHeader(values)
		}
	case HeaderCreator:
		if r.Version.Compare(Version020) >= 0 {
			s.Creators = parseMultiValuedHeader(values)
		}
	case HeaderLanguage:
		if r.Version.Compare(Version020) >= 0 {
			s.Languages = parseMultiValuedHeader(values)
		}
	case HeaderYear:
		if r.Version.Compare(Version020) >= 0 {
			s.Year, err = getUniqueHeaderAs(values, false, strconv.Atoi)
		}
	case HeaderComment:
		if r.Version.Compare(Version020) >= 0 {
			// FIXME: How do we deal with multiple values here?
			// An error for multiple different values seems wrong...
			// Maybe this should be multi valued? What about commas then?
			s.Comment = values[0]
		}
	default:
		s.ExtraHeaders.SetValues(key, values)
	}
	return
}

// TODO: Doc. If error is returned it is EncodingError.
func (r *Reader) UseEncoding(e encoding.Encoding) error {
	if e == r.Encoding {
		// nothing to be done
		return nil
	}

	h := r.Header.Copy()
	if r.Encoding != nil {
		if err := h.ApplyTransformer(r.Encoding.NewEncoder()); err != nil {
			return err
		}
	}
	r.Encoding.NewDecoder()
	if e != nil {
		if err := h.ApplyTransformer(e.NewDecoder()); err != nil {
			return err
		}
	}
	r.Header = h
	r.Encoding = e
	return nil
}

// scan reads the next line of input.
// If r.useBuffered is true this operation does not advance the underlying scanner and r.line will not change.
// Otherwise, the underlying scanner is advanced and r.line and r.line are updated accordingly.
func (r *Reader) scan() bool {
	r.line++
	if r.rescan {
		r.rescan = false
		return true
	}
	return r.s.Scan()
}

// unscan sets a flag in r such that the next call to r.scan will not advance the underlying scanner.
// This effectively causes scan to read the same line again.
func (r *Reader) unscan() {
	if r.line == 0 {
		panic("unscan called before scan.")
	}
	if r.rescan {
		panic("unscan called twice without scan.")
	}
	r.line--
	r.rescan = true
}
