package ultrastar

import (
	"bytes"
	"errors"
	"iter"
	"slices"
	"strconv"
	"strings"
	"time"
)

// These are the versions of the UltraStar file format explicitly supported by
// this package.
var (
	Version010 = MustParseVersion("0.1.0")
	Version020 = MustParseVersion("0.2.0")
	Version030 = MustParseVersion("0.3.0")
	Version100 = MustParseVersion("1.0.0")
	Version110 = MustParseVersion("1.1.0")
	Version120 = MustParseVersion("1.2.0")
	Version200 = MustParseVersion("2.0.0")
)

// These constants refer to known headers. The list of headers might not be
// exhaustive.
//
// For a detailed explanation of the headers, their availability, and their
// valid values see the [specification](https://usdx.eu/format/).
//
// For the sake of completeness, the list of constants also includes some
// non-standard headers that are commonly used.
const (
	HeaderVersion  = "VERSION"
	HeaderEncoding = "ENCODING"
	HeaderRelative = "RELATIVE"

	HeaderBPM      = "BPM"
	HeaderGap      = "GAP"
	HeaderVideoGap = "VIDEOGAP"

	HeaderPreviewStart    = "PREVIEWSTART"
	HeaderMedleyStart     = "MEDLEYSTART"
	HeaderMedleyEnd       = "MEDLEYEND"
	HeaderMedleyStartBeat = "MEDLEYSTARTBEAT"
	HeaderMedleyEndBeat   = "MEDLEYENDBEAT"
	HeaderCalcMedley      = "CALCMEDLEY"

	HeaderStart = "START"
	HeaderEnd   = "END"

	HeaderMP3             = "MP3"
	HeaderAudio           = "AUDIO"
	HeaderAudioURL        = "AUDIOURL"
	HeaderVocals          = "VOCALS"
	HeaderVocalsURL       = "VOCALSURL"
	HeaderInstrumental    = "INSTRUMENTAL"
	HeaderInstrumentalURL = "INSTRUMENTALURL"
	HeaderVideo           = "VIDEO"
	HeaderVideoURL        = "VIDEOURL"
	HeaderCover           = "COVER"
	HeaderCoverURL        = "COVERURL"
	HeaderBackground      = "BACKGROUND"
	HeaderBackgroundURL   = "BACKGROUNDURL"

	HeaderTitle      = "TITLE"
	HeaderArtist     = "ARTIST"
	HeaderRendition  = "RENDITION"
	HeaderYear       = "YEAR"
	HeaderGenre      = "GENRE"
	HeaderEdition    = "EDITION"
	HeaderLanguage   = "LANGUAGE"
	HeaderTags       = "TAGS"
	HeaderCreator    = "CREATOR"
	HeaderAuthor     = "AUTHOR" // alias for HeaderCreator
	HeaderAutor      = "AUTOR"  // alias for HeaderCreator
	HeaderProvidedBy = "PROVIDEDBY"
	HeaderComment    = "COMMENT"

	HeaderP1           = "P1"
	HeaderP2           = "P2"
	HeaderP3           = "P3"
	HeaderP4           = "P4"
	HeaderP5           = "P5"
	HeaderP6           = "P6"
	HeaderP7           = "P7"
	HeaderP8           = "P8"
	HeaderP9           = "P9"
	HeaderDuetSingerP1 = "DUETSINGERP1"
	HeaderDuetSingerP2 = "DUETSINGERP2"

	HeaderResolution = "RESOLUTION" // application-specific, USDX only
	HeaderNotesGap   = "NOTESGAP"   // application-specific, USDX only
)

// These are common error values when working with headers.
var (
	// ErrMultipleValues indicates that a Header contained multiple different values
	// for a single-valued header key.
	ErrMultipleValues = errors.New("multiple values")

	// ErrNoValue indicates that a single-valued header did not have a value.
	ErrNoValue = errors.New("no value")
)

// CanonicalHeaderKey returns the canonical version (upper-case version) of the
// specified key. If no canonical version of the key exists (e.g. if it contains
// invalid characters) an empty string is returned.
func CanonicalHeaderKey(key string) string {
	if strings.ContainsRune(key, ':') {
		return ""
	}
	return strings.ToUpper(strings.TrimSpace(key))
}

// Header represents the key-value pairs of an UltraStar file header.
//
// A single header key can have multiple values. Values of multivalued headers
// are not necessarily normalized. A nil-value, an empty array, and an absent
// key are all semantically equivalent.
//
// The keys should be in canonical form, as returned by CanonicalHeaderKey.
type Header map[string][]string

// Add adds the key, value pair to the header. It appends to any existing values
// associated with key. The key is case-insensitive; CanonicalHeaderKey is used
// to canonicalize the provided key.
func (h Header) Add(key, value string) {
	key = CanonicalHeaderKey(key)
	h[key] = append(h[key], value)
}

// Set sets the header entries associated with key to the single element value.
// It replaces any existing values associated with key. The key is
// case-insensitive; CanonicalHeaderKey is used to canonicalize the provided
// key. To use non-canonical keys, assign to the map directly.
func (h Header) Set(key, value string) {
	if value != "" {
		h[CanonicalHeaderKey(key)] = []string{value}
	} else {
		delete(h, CanonicalHeaderKey(key))
	}
}

// SetInt sets the header entries associated with key to the string
// representation of value. It replaces any existing values associated with key.
// The key is case-insensitive; CanonicalHeaderKey is used to canonicalize the
// provided key. To use non-canonical keys, assign to the map directly.
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

// SetFloat sets the header entries associated with key to the string
// representation of value. It replaces any existing values associated with key.
// The key is case-insensitive; CanonicalHeaderKey is used to canonicalize the
// provided key. To use non-canonical keys, assign to the map directly.
func (h Header) SetFloat(key string, value float64) {
	h.setFloat(key, value, false)
}

// setFloat sets the header entries associated with key to the string
// representation of value. It replaces any existing values associated with key.
// If comma is true the decimal point is replaced with a comma. The key is
// case-insensitive; CanonicalHeaderKey is used to canonicalize the provided
// key. To use non-canonical keys, assign to the map directly.
func (h Header) setFloat(key string, value float64, comma bool) {
	s := strconv.AppendFloat(nil, value, 'f', -1, 64)
	if comma {
		if i := bytes.IndexByte(s, '.'); i >= 0 {
			s[i] = ','
		}
	}
	h.Set(key, string(s))
}

// SetMultiValued sets the header entries associated with key to a value that
// encodes the given values as a multi-valued header. It replaces any existing
// values associated with key. The key is case-insensitive; CanonicalHeaderKey
// is used to canonicalize the provided key. To use non-canonical keys, assign
// to the map directly.
func (h Header) SetMultiValued(key string, values ...string) {
	h.Set(key, encodeMultiValue(values...))
}

// encodeMultiValue encodes a list of values to be used in a multi-valued
// header.
func encodeMultiValue(values ...string) string {
	i := 0
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		values[i] = strings.ReplaceAll(v, ",", ",,")
		i++
	}
	return strings.Join(values[:i], ",")
}

// SetValues replaces all header entries associated with key with the given
// values. The values slice is copied and empty elements are removed. The key is
// case-insensitive; CanonicalHeaderKey is used to canonicalize the provided
// key.
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

// Get gets a value associated with the given key. If there are no values
// associated with the key, Get returns "". It is case-insensitive;
// CanonicalHeaderKey is used to canonicalize the provided key. Get assumes that
// all keys are stored in canonical form. To use non-canonical keys, access the
// map directly.
//
// If a key has multiple values, it is undefined which value will be returned.
func (h Header) Get(key string) string {
	if h == nil {
		return ""
	}
	return getHeader(h[CanonicalHeaderKey(key)])
}

// getHeader returns the first non-empty value from values.
func getHeader(values []string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

// GetUnique gets the unique value associated with the given key. If there are
// multiple different non-empty values associated with the key, the error
// ErrMultipleValues is returned. If there is no value associated with the given
// key, an empty string and a nil error are returned. It is case-insensitive;
// CanonicalHeaderKey is used to canonicalize the provided key.
func (h Header) GetUnique(key string) (string, error) {
	return uniqueHeaderAs(h.Values(key), false, func(v string) (string, error) {
		return v, nil
	})
}

// GetMultiValued parses a multivalued header and returns all individual header
// values. See section 3.1 of the UltraStar file format specification for
// details. If there are no values associated with the key, GetMultiValued
// returns an empty slice. The key is case-insensitive; CanonicalHeaderKey is
// used to canonicalize the provided key. GetMultiValued assumes that all keys
// are stored in canonical form.
func (h Header) GetMultiValued(key string) iter.Seq[string] {
	return parseMultiValuedHeader(h.Values(key))
}

// Values returns all values associated with the given key. It is
// case-insensitive; CanonicalHeaderKey is used to canonicalize the provided
// key. To use non-canonical keys, access the map directly. The returned slice
// is not a copy.
func (h Header) Values(key string) []string {
	if h == nil {
		return nil
	}
	return h[CanonicalHeaderKey(key)]
}

// Has returns a bool indicating whether h contains a non-empty value for the
// given key. The key is case-insensitive; CanonicalHeaderKey is used to
// canonicalize the provided key. To use non-canonical keys, access the map
// directly.
func (h Header) Has(key string) bool {
	for _, v := range h.Values(key) {
		if v != "" {
			return true
		}
	}
	return false
}

// Del deletes the values associated with key. The key is case-insensitive;
// CanonicalHeaderKey is used to canonicalize the provided key.
func (h Header) Del(key string) {
	delete(h, CanonicalHeaderKey(key))
}

// Clean removes header values that are empty, invalid or consist only of
// whitespace. If a header key contains no values afterward, it is removed from
// h entirely. Clean uses CanonicalHeaderKey to canonicalize all keys
// (potentially merging values).
func (h Header) Clean() {
	cKeys := make(map[string]string)
	for key, values := range h {
		values = slices.DeleteFunc(values, func(v string) bool {
			return strings.TrimSpace(v) == ""
		})
		if len(values) == 0 {
			delete(h, key)
			continue
		}
		cKey := CanonicalHeaderKey(key)
		if cKey == "" {
			delete(h, key)
			continue
		}
		h[key] = values
		if key != cKey {
			cKeys[key] = cKey
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

// Clone creates a deep copy of h containing the same keys and values.
func (h Header) Clone() Header {
	if h == nil {
		return nil
	}

	// Find the total number of values.
	nv := 0
	for _, vv := range h {
		nv += len(vv)
	}
	sv := make([]string, nv) // shared backing array for headers' values
	h2 := make(Header, len(h))
	for k, vv := range h {
		if vv == nil {
			// Preserve nil values.
			h2[k] = nil
			continue
		}
		n := copy(sv, vv)
		h2[k] = sv[:n:n]
		sv = sv[n:]
	}
	return h2
}

// UniqueHeader returns the unique value associated with the given key in h. If
// there are multiple different non-empty values, the error ErrMultipleValues is
// returned. If no non-empty values exist for the given key, the empty string
// will be returned. If required is set to true, the error ErrNoValue is
// returned (otherwise the error is nil).
//
// The key is case-insensitive; CanonicalHeaderKey is used to canonicalize the
// provided key.
func UniqueHeader(h Header, key string, required bool) (string, error) {
	return uniqueHeaderAs(h.Values(key), required, func(v string) (string, error) { return v, nil })
}

// UniqueHeaderAs gets the unique value associated with the given key in h.
// Each non-empty value is transformed by conv. If there are multiple different
// values (after applying the transformation), the error ErrMultipleValues is
// returned. Any error returned from conv is returned directly. If no non-empty
// values exist for the given key, t will be the zero value of T. If required is
// set to true, the error ErrNoValue is returned (otherwise the error is nil).
//
// The key is case-insensitive; CanonicalHeaderKey is used to canonicalize the
// provided key.
func UniqueHeaderAs[T comparable](h Header, key string, required bool, conv func(string) (T, error)) (t T, err error) {
	return uniqueHeaderAs(h.Values(key), required, conv)
}

// uniqueHeaderAs works like [UniqueHeaderAs] but is passed the raw values
// of the header directly.
func uniqueHeaderAs[T comparable](values []string, required bool, conv func(string) (T, error)) (t T, err error) {
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

// parseMultiValuedHeader parses the given raw header values as a multivalued
// header. See section 3.1 of the UltraStar file format specification for
// details.
func parseMultiValuedHeader(vs []string) iter.Seq[string] {
	return func(yield func(string) bool) {
		if vs == nil {
			return
		}
		for _, rawValue := range vs {
			from := 0
			for rawValue != "" {
				j := strings.Index(rawValue[from:], ",")
				if j < 0 {
					break
				} else if len(rawValue) > j+1 && rawValue[j+1] == ',' {
					from += j + 2
					continue
				}
				frag := strings.TrimSpace(rawValue[:from+j])
				frag = strings.ReplaceAll(frag, ",,", ",")
				if frag != "" && !yield(frag) {
					return
				}
				rawValue = rawValue[from+j+1:]
				from = 0
			}
			rawValue = strings.TrimSpace(rawValue)
			rawValue = strings.ReplaceAll(rawValue, ",,", ",")
			if rawValue != "" && !yield(rawValue) {
				return
			}
		}
	}
}

// parseFloat converts v into a floating point value. If allowComma is true, the
// comma is recognized as a decimal separator.
func parseFloat(v string, allowComma bool) (float64, error) {
	if allowComma {
		v = strings.Replace(v, ",", ".", 1)
	}
	return strconv.ParseFloat(v, 64)
}

// parseDurationMilliseconds converts v into a duration by interpreting v as an
// integer indicating the number of milliseconds.
func parseDurationMilliseconds(v string) (time.Duration, error) {
	value, err := strconv.Atoi(v)
	return time.Duration(value) * time.Millisecond, err
}

// parseDurationSeconds converts v into a duration by interpreting v as a
// decimal number indicating the number of seconds.
//
// The comma is recognized as a decimal separator.
func parseDurationSeconds(v string) (time.Duration, error) {
	value, err := parseFloat(v, true)
	return time.Duration(value * float64(time.Second)), err
}
