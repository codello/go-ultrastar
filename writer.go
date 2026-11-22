package ultrastar

import (
	"errors"
	"io"
	"maps"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
)

// WriteSong serializes s into w using the specified version of the UltraStar
// file format. This is a convenience method for [Writer.WriteSong].
func WriteSong(w io.Writer, s *Song, version Version) error {
	wr := NewWriter(w, version)
	if err := wr.WriteSong(s); err != nil {
		return err
	}
	return wr.Close()
}

// A Writer implements the serialization of a [Song] into the UltraStar file format.
//
// A writer has to be created via the [NewWriter] function or the [Writer.Reset]
// method must be called before the first byte is written.
type Writer struct {
	// Version is the version of the file format to write. If Version is the zero
	// value, the Writer uses version 1.2.0.
	Version Version

	// Relative indicates that the writer will write notes in relative mode. This is
	// a legacy format that is not recommended anymore. Note that relative mode is
	// only supported for Version values below 1.0.0.
	Relative bool

	// CompareHeaders determines the order of header keys. It returns a negative
	// number if a comes before b, a positive number if a comes after b, and zero
	// otherwise. The keys are in canonical form. If CompareHeaders is nil, the
	// Writer uses the function [CompareHeaders].
	CompareHeaders func(a, b string) int

	wr     io.Writer // underlying writer
	buf    []byte    // current line, written to wr in a single call
	rel    [9]Beat   // relative offset for each voice
	voice  int       // current voice
	body   bool      // have any notes been written?
	closed bool
}

// NewWriter creates a new writer for UltraStar songs.
// The default settings aim to be compatible with most Karaoke games.
func NewWriter(wr io.Writer, version Version) *Writer {
	w := &Writer{Version: version}
	w.Reset(wr)
	return w
}

// Reset configures w to be reused, writing to wr.
// This method keeps the current writer's configuration.
func (w *Writer) Reset(wr io.Writer) {
	w.wr = wr
	clear(w.rel[:])
	w.voice = P1
	w.body = false
	w.closed = false
}

// version returns the version of the file format that w writes.
func (w *Writer) version() Version {
	if w.Version.IsZero() {
		return Version120
	}
	return w.Version
}

// relative determines if w is in relative mode.
func (w *Writer) relative() bool {
	return w.version().Major == 0 && w.Relative
}

// WriteSong writes the song s to w in the UltraStar txt format.
// If an error occurs, it is returned otherwise nil is returned.
//
// The Writer writes the headers of the song in the order that w.CompareHeaders
// determines.
func (w *Writer) WriteSong(song *Song) error {
	if len(song.Voices) > 9 {
		return errors.New("song has more than 9 voices")
	}
	if w.version().Major == 0 && len(song.Voices) > 2 {
		return errors.New("song has too many voices for version 0")
	}
	if w.closed {
		return errors.New("writer is already closed")
	}
	if w.body {
		return errors.New("header already written")
	}
	if err := w.writeHeader(w.songHeader(song)); err != nil {
		return err
	}
	if song.IsDuet() {
		// we want to include the leading P1 for duets
		w.VoiceChange()
	}
	for i, voice := range song.Voices {
		for _, n := range voice.Notes {
			if err := w.WriteNote(n, i); err != nil {
				return err
			}
		}
	}
	return nil
}

// songHeader returns the header that w writes for song. The fields of song take
// precedence over the values in song.Header: a standard header that a field
// does not set for w.Version is removed.
func (w *Writer) songHeader(song *Song) Header {
	version := w.version()
	h := song.Header.Clone()
	if h == nil {
		h = make(Header, 32)
	}
	h.Clean()

	h.Set(HeaderVersion, version.String())
	// The Writer always writes UTF-8.
	h.Del(HeaderEncoding)
	if w.relative() {
		h.Set(HeaderRelative, "YES")
	} else {
		h.Del(HeaderRelative)
	}

	h.Set(HeaderTitle, strings.TrimSpace(song.Title))
	h.Set(HeaderArtist, encodeMultiValue(song.Artist...))
	h.Set(HeaderRendition, strings.TrimSpace(song.Rendition))
	if song.Year != 0 {
		h.Set(HeaderYear, strconv.Itoa(song.Year))
	} else {
		h.Del(HeaderYear)
	}
	h.Set(HeaderGenre, encodeMultiValue(song.Genre...))
	h.Set(HeaderLanguage, encodeMultiValue(song.Language...))
	h.Set(HeaderEdition, encodeMultiValue(song.Edition...))
	h.Set(HeaderTags, encodeMultiValue(song.Tags...))
	h.Set(HeaderCreator, encodeMultiValue(song.Creator...))
	h.Set(HeaderProvidedBy, strings.TrimSpace(song.ProvidedBy))
	h.Set(HeaderComment, strings.TrimSpace(song.Comment))

	audio := strings.TrimSpace(song.Audio)
	h.Del(HeaderAudio)
	h.Del(HeaderMP3)
	if version.Major >= 1 {
		h.Set(HeaderAudio, audio)
	}
	if version.Major <= 1 {
		h.Set(HeaderMP3, audio)
	}
	h.Set(HeaderAudioURL, w.encodeURL(song.AudioURL))
	h.Set(HeaderVocals, strings.TrimSpace(song.Vocals))
	h.Set(HeaderVocalsURL, w.encodeURL(song.VocalsURL))
	h.Set(HeaderInstrumental, strings.TrimSpace(song.Instrumental))
	h.Set(HeaderInstrumentalURL, w.encodeURL(song.InstrumentalURL))
	h.Set(HeaderVideo, strings.TrimSpace(song.Video))
	h.Set(HeaderVideoURL, w.encodeURL(song.VideoURL))
	h.Set(HeaderCover, strings.TrimSpace(song.Cover))
	h.Set(HeaderCoverURL, w.encodeURL(song.CoverURL))
	h.Set(HeaderBackground, strings.TrimSpace(song.Background))
	h.Set(HeaderBackgroundURL, w.encodeURL(song.BackgroundURL))

	bpm := song.BPM
	if version.Major < 2 {
		bpm /= 4
	}
	h.Set(HeaderBPM, strconv.FormatFloat(float64(bpm), 'f', -1, 64))
	h.Set(HeaderGap, w.encodeDuration(song.Gap, false))
	h.Set(HeaderVideoGap, w.encodeDuration(song.VideoGap, true))
	h.Set(HeaderStart, w.encodeDuration(song.Start, true))
	h.Set(HeaderEnd, w.encodeDuration(song.End, false))
	h.Set(HeaderPreviewStart, w.encodeDuration(song.PreviewStart, true))
	h.Del(HeaderMedleyStart)
	h.Del(HeaderMedleyStartBeat)
	h.Del(HeaderMedleyEnd)
	h.Del(HeaderMedleyEndBeat)
	if song.MedleyStart != 0 && version.Major >= 2 {
		h.Set(HeaderMedleyStart, w.encodeDuration(song.MedleyStart, false))
	} else if song.MedleyStart != 0 {
		h.Set(HeaderMedleyStartBeat, strconv.Itoa(int(song.BPM.Beats(song.MedleyStart-song.Gap))))
	}
	if song.MedleyEnd != 0 && version.Major >= 2 {
		h.Set(HeaderMedleyEnd, w.encodeDuration(song.MedleyEnd, false))
	} else if song.MedleyEnd != 0 {
		h.Set(HeaderMedleyEndBeat, strconv.Itoa(int(song.BPM.Beats(song.MedleyEnd-song.Gap))))
	}

	for i := range 9 {
		key := "P" + strconv.Itoa(i+1)
		if i < len(song.Voices) {
			h.Set(key, strings.TrimSpace(song.Voices[i].Name))
		} else {
			h.Del(key)
		}
	}
	return h
}

// encodeDuration encodes the given duration to be used in a song header. The
// encoding may depend on w.Version. If legacySeconds is true, the duration will
// be encoded using a unit of seconds for versions before 2.0.0.
func (w *Writer) encodeDuration(duration time.Duration, legacySeconds bool) string {
	if duration == 0 {
		return ""
	}
	if legacySeconds && w.version().Major < 2 {
		return strconv.FormatFloat(duration.Seconds(), 'f', -1, 64)
	}
	return strconv.FormatInt(duration.Milliseconds(), 10)
}

// encodeURL encodes url to be used in a song header.
func (w *Writer) encodeURL(url *url.URL) string {
	if url == nil {
		return ""
	}
	return url.String()
}

// WriteHeader writes the specified header to the underlying writer. If w has
// already finished writing its header, a non-nil error will be returned.
//
// Header keys will be canonicalized by CanonicalHeaderKey and will be written
// in the order that w.CompareHeaders determines.
func (w *Writer) WriteHeader(h Header) error {
	if w.closed {
		return errors.New("writer is already closed")
	}
	if w.body {
		return errors.New("header already written")
	}
	h = h.Clone()
	h.Clean()
	return w.writeHeader(h)
}

// writeHeader writes h in the order that w.CompareHeaders determines. The keys
// of h must be in canonical form.
func (w *Writer) writeHeader(h Header) error {
	compare := w.CompareHeaders
	if compare == nil {
		compare = CompareHeaders
	}
	for _, key := range slices.SortedFunc(maps.Keys(h), compare) {
		for _, value := range h[key] {
			if err := w.writeRawHeader(key, value); err != nil {
				return err
			}
		}
	}
	return nil
}

// WriteRawHeader writes the specified header key and value verbatim to the
// underlying writer. If w has already finished writing its header, a non-nil
// error will be returned.
func (w *Writer) WriteRawHeader(key, value string) error {
	if w.closed {
		return errors.New("writer is already closed")
	}
	if w.body {
		return errors.New("header already written")
	}
	return w.writeRawHeader(key, value)
}

// writeRawHeader writes the given key and value as a header field. No further
// validation is performed.
func (w *Writer) writeRawHeader(key, value string) error {
	w.buf = append(w.buf[:0], '#')
	w.buf = append(w.buf, key...)
	w.buf = append(w.buf, ':')
	w.buf = append(w.buf, value...)
	w.buf = append(w.buf, '\n')
	_, err := w.wr.Write(w.buf)
	return err
}

// WriteNote writes n for the specified voice. If the voice differs from the
// voice of the previous note, a voice change is inserted. Depending on
// w.Relative, the note is adjusted to the current relative offset.
func (w *Writer) WriteNote(n Note, voice int) error {
	if w.closed {
		return errors.New("writer is already closed")
	}
	if voice < P1 || voice > P9 {
		panic("invalid voice change")
	}
	w.body = true
	if voice != w.voice {
		w.buf = append(w.buf[:0], 'P')
		w.buf = strconv.AppendInt(w.buf, int64(voice+1), 10)
		w.buf = append(w.buf, '\n')
		if _, err := w.wr.Write(w.buf); err != nil {
			return err
		}
		w.voice = voice
	}
	if w.relative() {
		n.Start -= w.rel[voice]
	}
	w.buf = append(w.buf[:0], byte(n.Type), ' ')
	w.buf = strconv.AppendInt(w.buf, int64(n.Start), 10)
	switch {
	case n.Type != NoteTypeEndOfPhrase:
		w.buf = append(w.buf, ' ')
		w.buf = strconv.AppendInt(w.buf, int64(n.Duration), 10)
		w.buf = append(w.buf, ' ')
		w.buf = strconv.AppendInt(w.buf, int64(n.Pitch), 10)
		w.buf = append(w.buf, ' ')
		w.buf = append(w.buf, n.Text...)
	case w.relative():
		w.buf = append(w.buf, ' ')
		w.buf = strconv.AppendInt(w.buf, int64(n.Start), 10)
		w.rel[voice] += n.Start
	}
	w.buf = append(w.buf, '\n')
	_, err := w.wr.Write(w.buf)
	return err
}

// VoiceChange registers a voice change.
// The next call to WriteNotes or WriteNote will write a voice change before the note,
// even if the voice didn't change compared to the previous note.
func (w *Writer) VoiceChange() {
	w.voice = -1
}

// Close writes the final "E" line of the song.
// Anything written to w or its underlying writer after this method returns
// will be ignored by programs reading the song.
//
// This method does not close the underlying writer of w.
func (w *Writer) Close() error {
	if w.closed {
		return errors.New("writer closed twice")
	}
	_, err := io.WriteString(w.wr, "E\n")
	w.closed = true
	return err
}
