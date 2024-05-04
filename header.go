package ultrastar

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/transform"
)

// TODO: Tests

// These are the versions of the UltraStar file format explicitly supported by this package.
var (
	Version010 = MustParseVersion("0.1.0")
	Version020 = MustParseVersion("0.2.0")
	Version030 = MustParseVersion("0.3.0")
	Version100 = MustParseVersion("1.0.0")
	Version110 = MustParseVersion("1.1.0")
	Version120 = MustParseVersion("1.2.0")
	Version200 = MustParseVersion("2.0.0")
)

// These constants refer to known headers.
// The list of headers might not be exhaustive.
//
// For a detailed explanation of the headers, their availability, and their valid values see the
// [specification](https://usdx.eu/format/).
//
// For the sake of completeness the list of constants also includes some non-standard headers that are commonly used.
const (
	HeaderVersion  = "VERSION"  // section 3.3
	HeaderEncoding = "ENCODING" // section 3.29
	HeaderRelative = "RELATIVE" // section 3.30

	HeaderBPM      = "BPM"      // section 3.4
	HeaderGap      = "GAP"      // section 3.11
	HeaderVideoGap = "VIDEOGAP" // section 3.12

	HeaderPreviewStart    = "PREVIEWSTART"    // section  3.15
	HeaderMedleyStart     = "MEDLEYSTART"     // section 3.16
	HeaderMedleyEnd       = "MEDLEYEND"       // section 3.16
	HeaderMedleyStartBeat = "MEDLEYSTARTBEAT" // section 3.17
	HeaderMedleyEndBeat   = "MEDLEYENDBEAT"   // section 3.17
	HeaderCalcMedley      = "CALCMEDLEY"      // section 3.18

	HeaderStart = "START" // section 3.14
	HeaderEnd   = "END"   // section 3.14

	HeaderMP3          = "MP3"          // section 3.6
	HeaderAudio        = "AUDIO"        // section 3.5
	HeaderVocals       = "VOCALS"       // section 3.10
	HeaderInstrumental = "INSTRUMENTAL" // section 3.10
	HeaderVideo        = "VIDEO"        // section 3.9
	HeaderCover        = "COVER"        // section 3.9
	HeaderBackground   = "BACKGROUND"   // section 3.9

	HeaderTitle      = "TITLE"      // section 3.7
	HeaderArtist     = "ARTIST"     // section 3.8
	HeaderYear       = "YEAR"       // section 3.19
	HeaderGenre      = "GENRE"      // section 3.20
	HeaderEdition    = "EDITION"    // section 3.22
	HeaderLanguage   = "LANGUAGE"   // section 3.21
	HeaderTags       = "TAGS"       // section 3.23
	HeaderCreator    = "CREATOR"    // section 3.26
	HeaderAuthor     = "AUTHOR"     // application-specific, alias for HeaderCreator
	HeaderAutor      = "AUTOR"      // application-specific, alias for HeaderCreator
	HeaderProvidedBy = "PROVIDEDBY" // section 3.27
	HeaderComment    = "COMMENT"    // section 3.28

	HeaderP1 = "P1" // section 3.24
	HeaderP2 = "P2" // section 3.24
	HeaderP3 = "P3" // section 3.24
	HeaderP4 = "P4" // section 3.24
	HeaderP5 = "P5" // section 3.24
	HeaderP6 = "P6" // section 3.24
	HeaderP7 = "P7" // section 3.24
	HeaderP8 = "P8" // section 3.24
	HeaderP9 = "P9" // section 3.24

	HeaderDuetSingerP1 = "DUETSINGERP1" // section 3.25
	HeaderDuetSingerP2 = "DUETSINGERP2" // section 3.25
	HeaderDuetSingerP3 = "DUETSINGERP3" // section 3.25
	HeaderDuetSingerP4 = "DUETSINGERP4" // section 3.25
	HeaderDuetSingerP5 = "DUETSINGERP5" // section 3.25
	HeaderDuetSingerP6 = "DUETSINGERP6" // section 3.25
	HeaderDuetSingerP7 = "DUETSINGERP7" // section 3.25
	HeaderDuetSingerP8 = "DUETSINGERP8" // section 3.25
	HeaderDuetSingerP9 = "DUETSINGERP9" // section 3.25

	HeaderResolution = "RESOLUTION" // application-specific, USDX only
	HeaderNotesGap   = "NOTESGAP"   // application-specific, USDX only
)

// These are common error values when working with headers.
var (
	// ErrMultipleValues indicates that a Header contained multiple different values for a single-valued header key.
	ErrMultipleValues = errors.New("multiple values")

	// ErrNoValue indicates that a single-valued header did not have a value.
	ErrNoValue = errors.New("no value")
)

// Header represents the key-value pairs of an UltraStar file header.
//
// A single header key can have multiple values.
// Values of multi-valued headers are not necessarily normalized.
// A nil-value, an empty array, and an absent key are all semantically equivalent.
//
// The keys should be in canonical form, as returned by CanonicalHeaderKey.
type Header map[string][]string

// Add adds the key, value pair to the header.
// It appends to any existing values associated with key.
// The key is case-insensitive; CanonicalHeaderKey is used to canonicalize the provided key.
func (h Header) Add(key, value string) {
	key = CanonicalHeaderKey(key)
	h[key] = append(h[key], value)
}

// Set sets the header entries associated with key to the single element value.
// It replaces any existing values associated with key.
// The key is case-insensitive; CanonicalHeaderKey is used to canonicalize the provided key.
// To use non-canonical keys, assign to the map directly.
func (h Header) Set(key, value string) {
	if value != "" {
		h[CanonicalHeaderKey(key)] = []string{value}
	} else {
		delete(h, CanonicalHeaderKey(key))
	}
}

// SetInt sets the header entries associated with key to the string representation of value.
// It replaces any existing values associated with key.
// The key is case-insensitive; CanonicalHeaderKey is used to canonicalize the provided key.
// To use non-canonical keys, assign to the map directly.
func (h Header) SetInt(key string, value int) {
	h.SetInt64(key, int64(value))
}

// SetInt64 sets the header entries associated with key to the string representation of value.
// It replaces any existing values associated with key.
// The key is case-insensitive; CanonicalHeaderKey is used to canonicalize the provided key.
// To use non-canonical keys, assign to the map directly.
func (h Header) SetInt64(key string, value int64) {
	h.Set(key, strconv.FormatInt(value, 10))
}

// SetFloat sets the header entries associated with key to the string representation of value.
// It replaces any existing values associated with key.
// The key is case-insensitive; CanonicalHeaderKey is used to canonicalize the provided key.
// To use non-canonical keys, assign to the map directly.
func (h Header) SetFloat(key string, value float64) {
	// TODO: Flag for comma-formatted floats?
	h.Set(key, strconv.FormatFloat(value, 'f', -1, 64))
}

// SetValues replaces all header entries associated with key with the given values.
// The values slice is copied and empty elements are removed.
// The key is case-insensitive; CanonicalHeaderKey is used to canonicalize the provided key.
func (h Header) SetValues(key string, values []string) {
	n := 0
	for _, v := range values {
		if v != "" {
			n++
		}
	}
	if n == 0 {
		h[CanonicalHeaderKey(key)] = nil
		return
	}
	vs := make([]string, n)
	j := 0
	for _, v := range values {
		if v != "" {
			vs[j] = v
			j++
		}
	}
	h[CanonicalHeaderKey(key)] = vs
}

// Get gets a value associated with the given key.
// If there are no values associated with the key, Get returns "".
// It is case-insensitive; CanonicalHeaderKey is used to canonicalize the provided key.
// Get assumes that all keys are stored in canonical form.
// To use non-canonical keys, access the map directly.
//
// If a key has multiple values it is undefined which value will be returned.
func (h Header) Get(key string) string {
	if h == nil {
		return ""
	}
	v := h[CanonicalHeaderKey(key)]
	if len(v) == 0 {
		return ""
	}
	return v[0]
}

// GetUnique gets the unique value associated with the given key.
// If there are multiple different non-empty values associated with the key, the error ErrMultipleValues is returned.
// If there is no value associated with the given key, an empty string and a nil error are returned.
// It is case-insensitive; CanonicalHeaderKey is used to canonicalize the provided key.
func (h Header) GetUnique(key string) (string, error) {
	return getUniqueHeaderAs(h.Values(key), false, func(v string) (string, error) {
		return v, nil
	})
}

// GetMultiValued parses a multi-valued header and returns all individual header values.
// See section 3.1 of the UltraStar file format specification for details.
// If there are no values associated with the key, GetMultiValued returns an empty slice.
// The key is case-insensitive; CanonicalHeaderKey is used to canonicalize the provided key.
// GetMultiValued assumes that all keys are stored in canonical form.
func (h Header) GetMultiValued(key string) []string {
	return parseMultiValuedHeader(h.Values(key))
}

// Values returns all values associated with the given key.
// It is case-insensitive; CanonicalHeaderKey is used to canonicalize the provided key.
// To use non-canonical keys, access the map directly.
// The returned slice is not a copy.
func (h Header) Values(key string) []string {
	if h == nil {
		return nil
	}
	return h[CanonicalHeaderKey(key)]
}

// Has returns a bool indicating whether h contains a non-empty value for the given key.
// The key is case-insensitive; CanonicalHeaderKey is used to canonicalize the provided key.
// To use non-canonical keys, access the map directly.
func (h Header) Has(key string) bool {
	for _, v := range h.Values(key) {
		if v != "" {
			return true
		}
	}
	return false
}

// Del deletes the values associated with key.
// The key is case-insensitive; CanonicalHeaderKey is used to canonicalize the provided key.
func (h Header) Del(key string) {
	delete(h, CanonicalHeaderKey(key))
}

// Clean removes header values that are empty or consist only of whitespace.
// If a header key contains no values afterward, it is removed from h entirely.
// Clean uses CanonicalHeaderKey to canonicalize all keys (potentially merging values).
func (h Header) Clean() {
	cKeys := make(map[string]string)
	for key, values := range h {
		slices.DeleteFunc(values, func(v string) bool {
			return strings.TrimSpace(v) == ""
		})
		if len(values) == 0 {
			delete(h, key)
		} else {
			clear(values[len(values):cap(values)])
			cKey := CanonicalHeaderKey(key)
			if key != cKey {
				cKeys[key] = cKey
			}
		}
	}
	for key, cKey := range cKeys {
		values := h[key]
		oldValues := h[cKey]
		if oldValues == nil {
			h[cKey] = values
		} else {
			h[cKey] = append(oldValues, values...)
		}
		delete(h, key)
	}
}

// Copy creates a deep copy of h containing the same keys and values.
func (h Header) Copy() Header {
	h2 := make(Header, len(h))
	for key, values := range h {
		h2[key] = append(h2[key], values...)
	}
	return h2
}

// ApplyTransformer applies t to all values of h.
// The transformation is performed in-place, neither h nor its value slices are copied.
// If an error occurs the transformation is aborted and the error is returned.
// If you need to restore the original values after an error, make a copy of h first.
func (h Header) ApplyTransformer(t transform.Transformer) (err error) {
	// FIXME: Does the encoding apply to keys as well?
	// TODO: What kind of error should be returned?
	for _, values := range h {
		for i, value := range values {
			if value, _, err = transform.String(t, value); err != nil {
				// TODO: Custom error type to indicate where the error occurred?
				return err
			}
			values[i] = value
		}
	}
	return nil
}

// GetUniqueHeaderAs gets the unique value associated with the given key in h.
// Each non-empty value is transformed by conv.
// If there are multiple different values (after applying the transformation), the error ErrMultipleValues is returned.
// Any error returned from conv is returned directly.
// If no non-empty values exist for the given key, t will be the zero value of T.
// If required is set to true, the error ErrNoValue is returned (otherwise the error is nil).
//
// The key is case-insensitive; CanonicalHeaderKey is used to canonicalize the provided key.
func GetUniqueHeaderAs[T comparable](h Header, key string, required bool, conv func(string) (T, error)) (t T, err error) {
	return getUniqueHeaderAs(h.Values(key), required, conv)
}

// getUniqueHeaderAs works like [GetUniqueHeaderAs] but is passed the raw values of the header directly.
func getUniqueHeaderAs[T comparable](values []string, required bool, conv func(string) (T, error)) (t T, err error) {
	var found bool
	for _, v := range values {
		if v == "" {
			continue
		}
		if tp, err := conv(v); err != nil {
			return t, err
		} else if found && t != tp {
			return t, ErrMultipleValues
		} else {
			t = tp
			found = true
		}
	}
	if required && !found {
		return t, ErrNoValue
	}
	return t, nil
}

// parseMultiValuedHeader parses the given raw header values as a multi-valued header.
// See section 3.1 of the UltraStar file format specification for details.
func parseMultiValuedHeader(vs []string) []string {
	if vs == nil {
		return nil
	}
	var values []string
	for _, rawValue := range vs {
		for _, value := range strings.Split(rawValue, ",") {
			// FIXME: What kinds of spaces should be trimmed?
			value = strings.TrimSpace(value)
			if value != "" {
				values = append(values, value)
			}
		}
	}
	return values
}

// parseFloat converts v into a floating point value.
// If allowComma is true, the comma is recognized as a decimal separator.
func parseFloat(v string, allowComma bool) (float64, error) {
	if allowComma {
		v = strings.ReplaceAll(v, ",", ".")
	}
	return strconv.ParseFloat(v, 64)
}

// parseDurationMilliseconds converts v into a duration by interpreting v
// as an integer indicating the number of milliseconds.
func parseDurationMilliseconds(v string) (time.Duration, error) {
	value, err := strconv.Atoi(v)
	return time.Duration(value) * time.Millisecond, err
}

// parseDurationSeconds converts v into a duration by interpreting v as a decimal
// number indicating the number of seconds.
//
// The comma is recognized as a decimal separator.
func parseDurationSeconds(v string) (time.Duration, error) {
	value, err := parseFloat(v, true)
	return time.Duration(value * float64(time.Second)), err
}

// CanonicalHeaderKey returns the canonical version (upper-case version) of the specified key.
// If no canonical version of the key exists (e.g. if it contains invalid characters) an empty string is returned.
func CanonicalHeaderKey(key string) string {
	if strings.ContainsRune(key, ':') {
		return ""
	}
	return strings.ToUpper(strings.TrimSpace(key))
}
