package ultrastar

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// Encodings is a map of known encodings that are recognized by the [Reader].
// You can add or remove encodings to configure the known encodings.
//
// Read and write access is not synchronized between multiple goroutines.
var Encodings = map[string]encoding.Encoding{
	"cp1250":       charmap.Windows1250,
	"cp-1250":      charmap.Windows1250,
	"windows1250":  charmap.Windows1250,
	"windows-1250": charmap.Windows1250,
	"cp1252":       charmap.Windows1252,
	"cp-1252":      charmap.Windows1252,
	"windows1252":  charmap.Windows1252,
	"windows-1252": charmap.Windows1252,
}

// SyntaxError records an error and the line number that could not be processed.
// Line numbers start at 1.
type SyntaxError struct {
	Line int
	Err  error
}

// Error returns the error string.
func (err *SyntaxError) Error() string {
	return fmt.Sprintf("syntax error at line %d: %v", err.Line, err.Err)
}

// Unwrap returns the underlying error.
func (err *SyntaxError) Unwrap() error {
	return err.Err
}

// HeaderError indicates an invalid Header value. The error type records the
// Header key as well as the underlying error.
//
// Methods and functions in the [ultrastar] package never return HeaderError's
// directly. Instead they return wrapped errors that can be tested against the
// HeaderError type using errors.Is and errors.As. HeaderError implements a
// special behavior for when used in these functions: If a Key is set on the
// target error, it only matches errors that relate to that key. Keys are
// normalized using CanonicalHeaderKey.
type HeaderError struct {
	Key string
	Err error
}

// Error returns the error string.
func (err *HeaderError) Error() string {
	return fmt.Sprintf("invalid value for header %s: %s", strconv.Quote("#"+err.Key), err.Err)
}

// Unwrap returns the underlying error.
func (err *HeaderError) Unwrap() error {
	return err.Err
}

// headerError is the internal type returned instead of a HeaderError. This type
// behaves like HeaderError and implements this equality via its Is and As
// methods.
type headerError HeaderError

// NewHeaderError returns a new error that can be compared to the HeaderError
// type like errors from the [ultrastar] package. The header key is normalized
// via CanonicalHeaderKey.
func NewHeaderError(key string, err error) error {
	return &headerError{CanonicalHeaderKey(key), err}
}

func (err *headerError) Error() string {
	return (*HeaderError)(err).Error()
}

func (err *headerError) Unwrap() error {
	return (*HeaderError)(err).Unwrap()
}

// Is implements equality checking for the HeaderError type. err is considered
// equivalent to a target *HeaderError if the key of the *HeaderError is non-empty
// and equal to err.Key. If err.Err is non-nil, the underlying errors need to be
// equal as well.
func (err *headerError) Is(target error) bool {
	//goland:noinspection GoTypeAssertionOnErrors
	hErr, ok := target.(*HeaderError)
	if !ok {
		return false
	}
	if hErr.Key != "" && hErr.Key != err.Key {
		return false
	}
	if err.Err != nil && errors.Is(hErr.Err, err.Err) {
		return true
	}
	return true
}

// As implements equality checking for the HeaderError type. err is considered
// equivalent to a target **HeaderError if either the underlying *HeaderError is
// nil or the underlying *HeaderError has a non-empty key that is equal to err.Key.
func (err *headerError) As(target any) bool {
	hErrPtr, ok := target.(**HeaderError)
	if !ok {
		return false
	}
	hErr := *hErrPtr
	if hErr != nil && hErr.Key != "" && hErr.Key != err.Key {
		return false
	}
	*hErrPtr = &HeaderError{err.Key, err.Err}
	return true
}

// ParseSong parses s into a Song. This is a convenience method for
// [Reader.ReadSong].
func ParseSong(s string) (*Song, error) {
	return ReadSong(strings.NewReader(s))
}

// ReadSong reads a Song from r. This is a convenience method for
// [Reader.ReadSong].
func ReadSong(r io.Reader) (*Song, error) {
	rd := &Reader{}
	if err := rd.Reset(r); err != nil {
		song, _ := rd.Song()
		return song, err
	}
	return rd.ReadSong()
}

// Reader implements a parser for the UltraStar file format. A Reader works in
// two phases:
//
//  1. When the reader is reset via Reader.Reset it reads the header of the
//     UltraStar file. Some header values (see below) are used to configure the
//     behavior of the reader.
//  2. The individual notes of a song are then read via
//     the Reader.ReadNote method. Alternatively the Reader.ReadSong method can be
//     used to construct a Song value.
//
// Some song headers are directly evaluated during the first phase of the
// Reader. Based on these values the corresponding fields of the reader are set.
// This determines the Reader behavior in its second phase. These headers are
// not passed on to Song values created by the Reader:
//
//	#VERSION
//	#RELATIVE
//	#ENCODING
//
// The zero value of a Reader is a valid reader, however Reader.Reset must be
// called before any read operation is performed. Alternatively the NewReader
// function can be used.
type Reader struct {
	// Header contains the raw header values read by a Reader. Header keys are
	// normalized using CanonicalHeaderKey. A Header is only valid until
	// [Reader.Reset] is called. If you need to access the headers afterward, you
	// must make a copy first.
	//
	// Modifications of Header values do not update the corresponding Reader fields
	// and updating Reader fields does not update Header values.
	Header Header

	// Version indicates the version of the UltraStar file format used by the
	// parser. This field determines version-dependent parsing behavior, such as the
	// unit of duration fields.
	//
	// This value is set by NewReader or Reader.Reset based on the #VERSION header.
	// It can also be set manually to change the behavior of the Read... methods.
	Version Version

	// Relative indicates whether the Reader is working in relative mode. In
	// relative mode, note times are interpreted as an offset relative to the
	// beginning of the phrase.
	//
	// This value is set by NewReader or Reader.Reset based on the RELATIVE header
	// if permitted by the Version. It can also be set manually to change the
	// behavior of the Read... methods.
	Relative bool

	// Encoding determines the encoding used by the ReadNote method.
	// The encoding is only applied to the note text, not the entire file.
	//
	// There are two options for changing the Encoding used by a Reader:
	//
	//   1. Updating the Encoding field will change the encoding for all future
	//      ReadNote operations.
	//   2. Using Reader.UseEncoding also updates the Encoding field but also
	//      re-encodes all Header keys and values. This can be useful if a song has
	//      been read in a wrong encoding, and you want to re-interpret the data in
	//      another encoding.
	//
	// This value is set by NewReader or Reader.Reset based on the #ENCODING header
	// if permitted by the Version. Detected encodings are determined by the
	// Encodings package variable. It can also be set manually to change the
	// behavior of the Read... methods.
	Encoding encoding.Encoding

	s      *bufio.Scanner // s reads from the underlying reader
	done   bool
	rescan bool
	line   int // current line number, set by scan
	voice  int // current voice, set by ReadNote
	rel    [9]Beat
}

// NewReader creates a new Reader instance and calls [Reader.Reset]. Any error
// during the reset is returned by this function. Regardless of error, a valid
// Reader is returned. It is valid to pass a nil argument for rd, but
// [Reader.Reset] must be called before any read operation is performed.
//
// See [Reader.Reset] for details on possible errors.
func NewReader(rd io.Reader) (*Reader, error) {
	r := &Reader{
		Header:  make(Header, 12),
		Version: Version030,
	}
	if rd == nil {
		return r, nil
	}
	return r, r.Reset(rd)
}

// Reset resets the internal state of r and reads a file header from rd. After
// this method returns, r.Header contains the header parsed from rd. This method
// parses and sets r.Version, r.Relative and r.Encoding from to the file header.
//
// There are multiple kinds of errors that can occur during the reset process:
//
//   - Any error from the underlying reader rd is returned as-is.
//   - An error wrapping one or more HeaderError-compatible errors if a header
//     recognized by Reader contains an invalid value.
func (r *Reader) Reset(rd io.Reader) error {
	r.s = bufio.NewScanner(newSkipPrefixReader(rd, []byte(utf8BOM)))
	r.s.Split(scanLines)
	r.done = false
	r.rescan = false
	r.line = 0
	r.voice = P1
	clear(r.rel[:])

	if r.Header == nil {
		r.Header = make(Header, 12)
	} else {
		clear(r.Header)
	}
	r.Relative = false
	r.Version = Version030
	r.Encoding = nil

	return r.readHeader()
}

// readHeader reads an UltraStar header from the underlying reader of r and
// configures r accordingly. This method fills r.Header. The returned error is
// either an error from the underlying reader or an error that can be tested
// against HeaderError if any of the headers recognized by the Reader type
// contain invalid values.
func (r *Reader) readHeader() (err error) {
	for ok := true; ok; {
		if ok, err = r.readHeaderLine(); err != nil {
			return err
		}
	}
	if v, err := uniqueHeaderAs(r.Header[HeaderVersion], true, ParseVersion); errors.Is(err, ErrNoValue) {
		r.Version = Version030
	} else if err != nil {
		return &headerError{HeaderVersion, err}
	} else {
		r.Version = v
	}
	if r.Version.Major == 0 {
		rel, relErr := uniqueHeaderAs(r.Header[HeaderRelative], false, func(v string) (bool, error) {
			return strings.EqualFold(v, "yes"), nil
		})
		if relErr == nil {
			r.Relative = rel
		} else {
			relErr = &headerError{HeaderRelative, relErr}
		}
		enc, encErr := uniqueHeaderAs(r.Header[HeaderEncoding], false, func(v string) (encoding.Encoding, error) {
			if enc, ok := Encodings[strings.ToLower(v)]; ok {
				return enc, nil
			} else {
				return nil, fmt.Errorf("unknown encoding %q", v)
			}
		})
		if encErr == nil {
			r.UseEncoding(enc)
		} else {
			encErr = &headerError{HeaderEncoding, encErr}
		}
		return errors.Join(encErr, relErr)
	}
	return nil
}

// readHeaderLine reads the next line from r.s, parses it as a header and stores
// the result in r.Header. The first return value indicates if a header line was
// read successfully. If the underlying reader generates an error, that error is
// returned.
func (r *Reader) readHeaderLine() (bool, error) {
	if !r.scan() {
		return false, r.s.Err()
	}
	line := bytes.TrimSpace(r.s.Bytes())
	if len(line) == 0 {
		return true, nil
	}
	if line[0] != '#' {
		r.unscan()
		return false, nil
	}
	if len(line) == 1 {
		// we skip lines that contain only the # sign
		return true, nil
	}
	index := bytes.IndexByte(line, ':')
	if index < 0 {
		key := CanonicalHeaderKey(string(line))
		if !r.Header.Has(key) {
			r.Header[key] = nil
		}
		return true, nil
	}
	r.Header.Add(
		string(bytes.TrimSpace(line[1:index])),
		string(bytes.TrimSpace(line[index+1:])),
	)
	return true, nil
}

// ReadNote reads the next note line from the input. This method understands and
// interprets any voice changes preceding the note line. If the underlying
// reader returns an error, the error is returned as-is. If the note line cannot
// be parsed, an error wrapping SyntaxError will be returned. If the song has
// been read completely (either to EOF or until an end tag has been read) the
// returned error will be io.EOF.
//
// In case of a SyntaxError, the returned note may contain partial parse results.
func (r *Reader) ReadNote() (note Note, voice int, err error) {
	if r.done {
		return Note{}, 0, io.EOF
	}

	// There can be multiple voice changes without a single note.
	// Only the last voice change is significant.
	for {
		if !r.scan() {
			if r.s.Err() == nil {
				return Note{}, -1, io.EOF
			}
			return Note{}, r.voice, r.s.Err()
		}
		line := r.s.Bytes()
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
			r.done = true
			return Note{}, -1, io.EOF
		case 'P':
			voice, err = strconv.Atoi(string(bytes.TrimSpace(trimmed[1:])))
			if err != nil || voice <= 0 {
				return Note{}, -1, &SyntaxError{r.line, errors.New("invalid voice change")}
			}
			voice--
			r.voice = voice
			continue
		}
		note, err = r.parseNote(line)
		if err != nil {
			return note, r.voice, &SyntaxError{r.line, fmt.Errorf("invalid note: %w", err)}
		}
		return note, r.voice, nil
	}
}

// parseNote parses a note line into a Note value. If an error is encountered
// the returned Note contains partial parse results.
func (r *Reader) parseNote(line []byte) (Note, error) {
	note := Note{}
	if len(line) == 0 {
		return note, errors.New("invalid note type")
	}
	note.Type = NoteType(line[0])
	line = line[1:]
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

// nextField returns the firs contiguous run of non-whitespace runes from bs as
// a string and a sub-slice containing the remainder of bs. This method ignores
// leading whitespace characters. The remainder may start with a whitespace
// character.
func nextField(bs []byte) (field string, remainder []byte) {
	bs = bytes.TrimLeftFunc(bs, func(r rune) bool {
		return unicode.Is(unicode.White_Space, r)
	})
	i := bytes.IndexFunc(bs, func(r rune) bool {
		return unicode.Is(unicode.White_Space, r)
	})
	if i < 0 {
		i = len(bs)
	}
	return string(bs[:i]), bs[i:]
}

// ReadSong parses a [Song] from r. If the song ends with an end tag (a line
// containing only the letter 'E') r may not be read until the end.
//
// If err is non-nil this method returns a partial result up to the point where
// the error was encountered. This method may return errors wrapping HeaderError
// (if a header value is invalid), SyntaxError (if a note line cannot be parsed)
// or errors from the underlying reader.
//
// This method normalizes the song's voices by removing voices that have neither
// a name nor notes.
func (r *Reader) ReadSong() (*Song, error) {
	return r.readSong(false)
}

// readSong parses r.Header into the song and then reads as much of the input as
// possible to fill the voices of the song with notes. If force is true, this
// method collects syntax errors and tries its best to parse any note line it
// can find. The return value is a song and a wrapped collection of errors that
// occurred during parsing.
func (r *Reader) readSong(force bool) (*Song, error) {
	var errs []error
	song, err := r.Song()
	if err != nil && force {
		errs = append(errs, err)
	} else if err != nil {
		return song, err
	}
	song.Voices = slices.Grow(song.Voices, 9-len(song.Voices))[:max(9, len(song.Voices))]
	var note Note
	var voice int
	var sErr *SyntaxError
	for {
		note, voice, err = r.ReadNote()
		if errors.As(err, &sErr) && force {
			errs = append(errs, err)
		} else if err != nil {
			break
		}
		song.Voices[voice].AppendNotes(note)
	}
	song.Voices = slices.DeleteFunc(song.Voices, func(voice *Voice) bool {
		return voice.Name == "" && voice.IsEmpty()
	})
	for _, voice := range song.Voices {
		voice.SortNotes()
	}
	if errors.Is(err, io.EOF) {
		return song, errors.Join(errs...)
	}
	if force {
		return song, err
	} else {
		errs = append(errs, err)
		return song, errors.Join(errs...)
	}
}

// Song creates a Song from the headers parsed by r. This method does not
// advance the reader or attempt to read any notes. This method only parses
// known song headers.
//
// If the headers contain voice names, these will be set on the corresponding
// elements of song.Voices.
func (r *Reader) Song() (*Song, error) {
	errs := make([]error, len(r.Header))
	song := &Song{Header: make(Header, len(r.Header)-4)}

	// Other headers depend on GAP and BPM, so these need to be processed first
	if err := r.setSongHeader(song, HeaderBPM, r.Header[HeaderBPM]); err != nil {
		errs = append(errs, &headerError{HeaderBPM, err})
	}
	if err := r.setSongHeader(song, HeaderGap, r.Header[HeaderGap]); err != nil {
		errs = append(errs, &headerError{HeaderGap, err})
	}
	for key, values := range r.Header {
		if key == HeaderBPM || key == HeaderGap {
			continue
		}
		if err := r.setSongHeader(song, key, values); err != nil {
			errs = append(errs, &headerError{key, err})
		}
	}
	song.Voices = make([]*Voice, 9)
	for i := range song.Voices {
		name := r.Header.Get("P" + strconv.Itoa(i+1))
		if name == "" && r.Version.Major < 1 {
			name = r.Header.Get("DUETSINGERP" + strconv.Itoa(i+1))
		}
		song.Voices[i] = &Voice{Name: name}
	}
	i := slices.IndexFunc(song.Voices, func(voice *Voice) bool {
		return voice.Name != ""
	})
	song.Voices = song.Voices[:i+1]

	return song, errors.Join(errs...)
}

// setSongHeader sets the header with the specified key in s. If the header is a
// known header it is parsed from the specified values, and the corresponding
// field is set on s. Headers that have no special behavior assigned in the
// version indicated by r.Version are copied as specified into s.
func (r *Reader) setSongHeader(s *Song, key string, values []string) (err error) {
	switch key {
	case HeaderVersion, HeaderP1, HeaderP2, HeaderP3, HeaderP4, HeaderP5, HeaderP6, HeaderP7, HeaderP8, HeaderP9:
		// These headers are recognized by the Reader and are not passed to s.
	case HeaderEncoding, HeaderRelative, HeaderDuetSingerP1, HeaderDuetSingerP2:
		// These headers are recognized by the Reader and are not passed to s.
		if r.Version.Compare(Version100) >= 0 {
			s.Header.SetValues(key, values)
		}

	case HeaderTitle:
		s.Title = getHeader(values)
	case HeaderArtist:
		fmt.Printf("Artists: %s", values)
		s.Artist = slices.Collect(parseMultiValuedHeader(values))
	case HeaderRendition:
		s.Rendition = getHeader(values)
	case HeaderYear:
		s.Year, err = uniqueHeaderAs(values, false, strconv.Atoi)
	case HeaderGenre:
		s.Genre = slices.Collect(parseMultiValuedHeader(values))
	case HeaderLanguage:
		s.Language = slices.Collect(parseMultiValuedHeader(values))
	case HeaderEdition:
		s.Edition = slices.Collect(parseMultiValuedHeader(values))
	case HeaderTags:
		s.Tags = slices.Collect(parseMultiValuedHeader(values))
	case HeaderCreator:
		s.Creator = slices.Collect(parseMultiValuedHeader(values))
	case HeaderAuthor, HeaderAutor:
		if len(s.Creator) == 0 {
			s.Creator = slices.Collect(parseMultiValuedHeader(values))
		}
	case HeaderProvidedBy:
		s.ProvidedBy = getHeader(values)
	case HeaderComment:
		s.Comment = getHeader(values)

	case HeaderBPM:
		s.BPM, err = uniqueHeaderAs(values, true, func(v string) (BPM, error) {
			f, err := parseFloat(v, r.Version.Major < 2)
			b := BPM(f) * 4
			if !b.IsValid() {
				return b, fmt.Errorf("invalid BPM value: %f", b)
			}
			return b, err
		})
	case HeaderGap:
		if r.Version.Major < 2 {
			s.Gap, err = uniqueHeaderAs(values, false, func(v string) (time.Duration, error) {
				f, err := parseFloat(v, true)
				return time.Duration(f * float64(time.Millisecond)), err
			})
		} else {
			s.Gap, err = uniqueHeaderAs(values, false, parseDurationMilliseconds)
		}
	case HeaderVideoGap:
		if r.Version.Major >= 2 {
			s.VideoGap, err = uniqueHeaderAs(values, false, parseDurationMilliseconds)
		} else {
			s.VideoGap, err = uniqueHeaderAs(values, false, parseDurationSeconds)
		}
	case HeaderStart:
		if r.Version.Major >= 2 {
			s.Start, err = uniqueHeaderAs(values, false, parseDurationMilliseconds)
		} else {
			s.Start, err = uniqueHeaderAs(values, false, parseDurationSeconds)
		}
	case HeaderEnd:
		if r.Version.Major >= 2 {
			s.End, err = uniqueHeaderAs(values, false, parseDurationMilliseconds)
		} else {
			s.End, err = uniqueHeaderAs(values, false, func(s string) (time.Duration, error) {
				f, err := parseFloat(s, true)
				return time.Duration(f * float64(time.Millisecond)), err
			})
		}
	case HeaderPreviewStart:
		if r.Version.Major >= 2 {
			s.PreviewStart, err = uniqueHeaderAs(values, false, parseDurationMilliseconds)
		} else {
			s.PreviewStart, err = uniqueHeaderAs(values, false, parseDurationSeconds)
		}
	case HeaderMedleyStart:
		if r.Version.Major < 2 {
			s.Header.SetValues(key, values)
		} else {
			s.MedleyStart, err = uniqueHeaderAs(values, false, parseDurationMilliseconds)
		}
	case HeaderMedleyEnd:
		if r.Version.Major < 2 {
			s.Header.SetValues(key, values)
		} else {
			s.MedleyEnd, err = uniqueHeaderAs(values, false, parseDurationMilliseconds)
		}
	case HeaderMedleyStartBeat:
		if r.Version.Major < 2 {
			if st, pErr := uniqueHeaderAs(values, true, strconv.Atoi); pErr != nil && !errors.Is(pErr, ErrNoValue) {
				err = pErr
			} else if pErr == nil {
				s.MedleyStart = s.BPM.Duration(Beat(st)) + s.Gap
			}
		} else {
			s.Header.SetValues(key, values)
		}
	case HeaderMedleyEndBeat:
		if r.Version.Major < 2 {
			if e, pErr := uniqueHeaderAs(values, true, strconv.Atoi); pErr != nil && !errors.Is(pErr, ErrNoValue) {
				err = pErr
			} else if pErr == nil {
				s.MedleyEnd = s.BPM.Duration(Beat(e)) + s.Gap
			}
		} else {
			s.Header.SetValues(key, values)
		}

	case HeaderMP3, HeaderAudio:
		s.Audio = getHeader(values)
	case HeaderAudioURL:
		s.AudioURL, err = url.Parse(getHeader(values))
	case HeaderVocals:
		s.Vocals = getHeader(values)
	case HeaderVocalsURL:
		s.VocalsURL, err = url.Parse(getHeader(values))
	case HeaderInstrumental:
		s.Instrumental = getHeader(values)
	case HeaderInstrumentalURL:
		s.InstrumentalURL, err = url.Parse(getHeader(values))
	case HeaderVideo:
		s.Video = getHeader(values)
	case HeaderVideoURL:
		s.VideoURL, err = url.Parse(getHeader(values))
	case HeaderCover:
		s.Cover = getHeader(values)
	case HeaderCoverURL:
		s.CoverURL, err = url.Parse(getHeader(values))
	case HeaderBackground:
		s.Background = getHeader(values)
	case HeaderBackgroundURL:
		s.BackgroundURL, err = url.Parse(getHeader(values))

	default:
		s.Header.SetValues(key, values)
	}
	return
}

// Line returns the number of lines that have already been processed by r. Use
// this method after a call to ReadNote to get the line number of the note line.
func (r *Reader) Line() int {
	return r.line
}

// UseEncoding sets r.Encoding to the specified encoding. All future read
// operations of r will use the new encoding. This method decodes all keys and
// values of r.Header in the new encoding. If r had an encoding set prior to
// calling this method, header values are first re-encoded into that encoding.
//
// Use this method to rectify r having read its headers in a wrong encoding. To
// just set the encoding for future read operations set r.Encoding directly.
//
// A nil encoding is understood to be UTF-8.
func (r *Reader) UseEncoding(e encoding.Encoding) {
	if e == r.Encoding {
		// nothing to be done
		return
	}

	var t transform.Transformer
	if r.Encoding != nil {
		t = encoding.ReplaceUnsupported(r.Encoding.NewEncoder())
	}
	if e != nil {
		if t != nil {
			t = transform.Chain(t, e.NewDecoder())
		} else {
			t = e.NewDecoder()
		}
	}

	// If both e and r.Encoding behave correctly. there should not be any errors.
	// encoding.ReplaceUnsupported takes care of any non-encodable runes in e and an
	// encoding.Decoder should not return errors for data that is not of that
	// encoding.
	for key, values := range r.Header {
		for i, value := range values {
			values[i], _, _ = transform.String(t, value)
		}
		newKey, _, _ := transform.String(t, key)
		if key != newKey {
			r.Header[newKey] = values
			delete(r.Header, key)
		}
	}
	r.Encoding = e
}

// scan reads the next line of input. If this scan call is preceded by a call to
// unscan, no new data will be read from the underlying reader.
func (r *Reader) scan() bool {
	if r.rescan {
		r.line++
		r.rescan = false
		return true
	}
	for r.s.Scan() {
		r.line++
		// ignore empty lines
		if len(r.s.Bytes()) > 0 {
			return true
		}
	}
	return false
}

// unscan reverts the last call to scan. The next time scan is called, no new
// data will be read from the underlying reader. It is an error to call unscan
// multiple times without calling scan in between.
//
// The unscan method is used to implement a lookahead of 1 line.
func (r *Reader) unscan() {
	if r.line == 0 {
		panic("unscan called before scan")
	}
	if r.rescan {
		panic("unscan called twice without scan")
	}
	r.line--
	r.rescan = true
}

// scanLines is a split function for a bufio.Scanner that returns each line of
// text, stripped of any trailing end-of-line marker. The returned line may be
// empty. The end-of-line marker is a single newline, a single carriage return,
// or a carriage return followed by a newline. The last non-empty line of input
// will be returned even if it has no newline.
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

// utf8BOM is the UTF8 byte order mark.
const utf8BOM = "\ufeff"

// skipPrefixReader implements an io.Reader that drops a known prefix from the
// input (if it is present). The first call to the Read method blocks until
// enough characters for the prefix have been read.
type skipPrefixReader struct {
	rd     io.Reader
	prefix []byte
	buf    []byte
	done   bool
	err    error
}

// newSkipPrefixReader creates a new skipPrefixReader that skips the given
// prefix. The reader will read its data from rd.
func newSkipPrefixReader(rd io.Reader, prefix []byte) *skipPrefixReader {
	return &skipPrefixReader{rd: rd, prefix: prefix, buf: make([]byte, 0, len(prefix))}
}

func (r *skipPrefixReader) Read(p []byte) (n int, err error) {
	if r.done {
		if len(r.buf) > 0 {
			n = copy(p, r.buf)
			r.buf = r.buf[n:]
			return n, nil
		}
		if r.err != nil {
			return 0, r.err
		}
		return r.rd.Read(p)
	}
	if len(p) == 0 {
		return 0, nil
	}
	r.done = true
	lastCopy := 0
	for r.err == nil && len(r.buf) < len(r.prefix) {
		n, r.err = r.rd.Read(p)
		lastCopy = copy(r.buf[len(r.buf):cap(r.buf)], p[:n])
		r.buf = r.buf[:len(r.buf)+lastCopy]
		// if !bytes.Equal(r.buf[len(r.buf)-lastCopy:], r.prefix[len(r.buf)-lastCopy:len(r.buf)]) {
		// 	break
		// }
	}
	if bytes.HasPrefix(r.buf, r.prefix) {
		copy(p, p[lastCopy:n])
		n -= lastCopy
		r.buf = r.buf[:0]
		return n, r.err
	}
	// we have valid data in buf and p that needs to be concatenated
	if len(r.buf)+n-lastCopy <= len(p) { // we will be able to return all data
		err = r.err
	}
	if n-lastCopy == 0 { // all the data is in r.buf, none in p.
		n = copy(p, r.buf)
		r.buf = r.buf[n:]
		return n, err
	}
	// r.buf is filled, and there is data in p
	tmp := make([]byte, max(0, n+len(r.buf)-lastCopy-len(p)))
	copy(tmp, p[n-len(tmp):n])
	copy(p[len(r.buf):], p[lastCopy:n])
	copy(p, r.buf)
	copy(r.buf, tmp)
	r.buf = r.buf[:len(tmp)]
	return n, err
}
