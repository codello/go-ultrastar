package ultrastar

import (
	"cmp"
	"io"
	"slices"
	"strconv"
)

// WriteSong serializes s into w.
// This is a convenience method for [Format.WriteSong].
func WriteSong(w io.Writer, s *Song) error {
	return NewWriter(w, Version120).WriteSong(s)
}

func WriteSongV2(w io.Writer, s *Song) error {
	return NewWriter(w, Version200).WriteSong(s)
}

// A Writer implements the serialization of a [Song] into the UltraStar file format.
type Writer struct {
	// FieldSeparator is a character used to separate fields in note line and line breaks.
	// This should only be set to a space or tab.
	//
	// Characters other than space or tab may or may not work and
	// will most likely result in invalid songs.
	FieldSeparator rune

	// TODO: Add config which multi-valued headers should be joined

	// Version is the version of the file format to write.
	Version Version

	// Relative indicates that the writer will write notes in relative mode.
	// This is a legacy format that is not recommended anymore.
	Relative bool

	// CommaFloat indicates that floating point values should use a comma as decimal separator.
	CommaFloat bool

	// TODO: Allow customization the order of tags

	wr    io.Writer // underlying writer
	rel   []Beat    // relative offset for each voice
	voice int       // current voice
}

// NewWriter creates a new writer for UltraStar songs.
// The default settings aim to be compatible with most Karaoke games.
func NewWriter(wr io.Writer, version Version) *Writer {
	w := &Writer{
		Version:        version,
		FieldSeparator: ' ',
		Relative:       false,
		CommaFloat:     false,
	}
	w.Reset(wr)
	return w
}

// Reset configures w to be reused, writing to wr.
// This method keeps the current writer's configuration.
func (w *Writer) Reset(wr io.Writer) {
	w.wr = wr
	w.rel = make([]Beat, 9)
	w.voice = P1
	// FIXME: Maybe we want to set w.voice to -1 to always include the first P1?
}

// WriteSong writes the song s to w in the UltraStar txt format.
// If an error occurs it is returned, otherwise nil is returned.
func (w *Writer) WriteSong(s *Song) error {
	if w.Version.IsZero() {
		w.Version = Version030
	}
	h := songHeader(s, w.Version)
	h.Set(HeaderVersion, w.Version.String())
	if w.Relative {
		h.Set(HeaderRelative, "YES")
	} else {
		h.Del(HeaderRelative)
	}
	if err := w.WriteHeader(h); err != nil {
		return err
	}
	if s.IsDuet() {
		// we want to include the leading P1 for duets
		w.VoiceChange()
	}
	for i, voice := range s.Voices {
		for _, n := range voice.Notes {
			if err := w.WriteNote(n, i); err != nil {
				return err
			}
		}
	}
	return nil
}

func songHeader(s *Song, v Version) Header {
	h := s.ExtraHeaders.Copy()
	if v.Compare(Version110) >= 0 {
		h[HeaderAudio] = s.Audio
	}
	if v.Compare(Version200) <= 0 {
		h[HeaderMP3] = s.Audio
	}
	// TODO: DO we want to overwrite custom values, even if the song values are empty?
	h[HeaderVocals] = s.Vocals
	h[HeaderInstrumental] = s.Instrumental
	h[HeaderVideo] = s.Video
	h[HeaderCover] = s.Cover
	h[HeaderBackground] = s.Background

	// FIXME: Maybe make these special set... methods private?
	h.SetFloat(HeaderBPM, float64(s.BPM))
	if s.Gap != 0 {
		h.SetInt64(HeaderGap, s.Gap.Milliseconds())
	}
	if s.VideoGap != 0 && v.Compare(Version200) >= 0 {
		h.SetInt64(HeaderVideoGap, s.VideoGap.Milliseconds())
	} else if s.VideoGap != 0 {
		h.SetFloat(HeaderVideoGap, s.VideoGap.Seconds())
	}
	if s.Start != 0 && v.Compare(Version200) >= 0 {
		h.SetInt64(HeaderStart, s.Start.Milliseconds())
	} else if s.Start != 0 {
		h.SetFloat(HeaderStart, s.Start.Seconds())
	}
	if s.End != 0 {
		h.SetInt64(HeaderEnd, s.End.Milliseconds())
	}
	if s.NoAutoMedley {
		// FIXME: Is this correct?
		h.Set(HeaderCalcMedley, "no")
	}
	if s.PreviewStart != 0 && v.Compare(Version200) >= 0 {
		h.SetInt64(HeaderPreviewStart, s.PreviewStart.Milliseconds())
	} else if s.PreviewStart != 0 {
		h.SetFloat(HeaderPreviewStart, s.PreviewStart.Seconds())
	}
	if s.MedleyStart != 0 && v.Compare(Version200) >= 0 {
		h.SetInt64(HeaderMedleyStart, s.MedleyStart.Milliseconds())
	} else if s.MedleyStart != 0 {
		h.SetInt(HeaderMedleyStartBeat, int(s.BPM.Beats(s.MedleyStart)))
	}
	if s.MedleyEnd != 0 && v.Compare(Version200) >= 0 {
		h.SetInt64(HeaderMedleyEnd, s.MedleyEnd.Milliseconds())
	} else if s.MedleyEnd != 0 {
		h.SetInt(HeaderMedleyEndBeat, int(s.BPM.Beats(s.MedleyEnd)))
	}
	// FIXME: Should we overwrite values even if the field is empty?
	if s.Title != "" {
		h.Set(HeaderTitle, s.Title)
	}
	// FIXME: Should we overwrite values even if the field is empty?
	h.Set(HeaderArtist, s.Artist)
	h[HeaderGenre] = s.Genres
	h[HeaderEdition] = s.Editions
	h[HeaderCreator] = s.Creators
	h[HeaderLanguage] = s.Languages
	h.SetInt(HeaderYear, s.Year)
	h.Set(HeaderComment, s.Comment)
	return h
}

type keyValues struct {
	Key    string
	Index  int
	Values []string
}

var headers010 = []string{HeaderVersion, HeaderTitle, HeaderArtist, HeaderMP3, HeaderBPM}
var headers020 = []string{HeaderVersion, HeaderEncoding, HeaderTitle, HeaderArtist, HeaderMP3, HeaderBPM, HeaderGap, HeaderCover, HeaderBackground, HeaderVideo, HeaderVideoGap, HeaderGenre, HeaderEdition, HeaderCreator, HeaderLanguage, HeaderYear, HeaderStart, HeaderEnd, HeaderPreviewStart, HeaderMedleyStartBeat, HeaderMedleyEndBeat, HeaderCalcMedley, HeaderComment, HeaderRelative}
var headers030 = headers020
var headers100 = []string{HeaderVersion, HeaderTitle, HeaderArtist, HeaderMP3, HeaderBPM, HeaderGap, HeaderCover, HeaderBackground, HeaderVideo, HeaderVideoGap, HeaderGenre, HeaderEdition, HeaderCreator, HeaderLanguage, HeaderYear, HeaderStart, HeaderEnd, HeaderPreviewStart, HeaderMedleyStartBeat, HeaderMedleyEndBeat, HeaderCalcMedley, HeaderComment}
var headers110 = []string{HeaderVersion, HeaderTitle, HeaderArtist, HeaderMP3, HeaderAudio, HeaderVocals, HeaderInstrumental, HeaderBPM, HeaderGap, HeaderCover, HeaderBackground, HeaderVideo, HeaderVideoGap, HeaderGenre, HeaderEdition, HeaderTags, HeaderCreator, HeaderLanguage, HeaderYear, HeaderStart, HeaderEnd, HeaderPreviewStart, HeaderMedleyStartBeat, HeaderMedleyEndBeat, HeaderCalcMedley, HeaderComment, HeaderProvidedBy}

func (w *Writer) WriteHeader(h Header) error {
	// FIXME: Should we return an error if notes have already been written?
	var standardHeaders []string
	switch {
	case w.Version.LessThan(Version020):
		standardHeaders = headers010
	case w.Version.LessThan(Version030):
		standardHeaders = headers020
	case w.Version.LessThan(Version100):
		standardHeaders = headers030
	case w.Version.LessThan(Version110):
		standardHeaders = headers100
	case w.Version.LessThan(Version120):
		standardHeaders = headers110
	default:
		standardHeaders = headers110
	}
	kvs := make([]keyValues, 0, len(h))
	for key, values := range h {
		kvs = append(kvs, keyValues{
			key,
			slices.Index(standardHeaders, key),
			values,
		})
	}
	slices.SortFunc(kvs, func(kv1, kv2 keyValues) int {
		if kv1.Index == kv2.Index {
			return cmp.Compare(kv1.Key, kv2.Key)
		} else if kv1.Index < 0 {
			return 1
		} else if kv2.Index < 0 {
			return -1
		} else {
			return cmp.Compare(kv1.Index, kv2.Index)
		}
	})
	for _, kv := range kvs {
		for _, vv := range kv.Values {
			if err := w.WriteHeaderLine(kv.Key, vv); err != nil {
				return err
			}
		}
	}
	return nil
}

// WriteHeaderLine writes a single header.
// Neither the header key nor the value are validated or normalized, they are written as-is.
func (w *Writer) WriteHeaderLine(key string, value string) error {
	for _, s := range []string{"#", key, ":", value, "\n"} {
		if _, err := io.WriteString(w.wr, s); err != nil {
			return err
		}
	}
	return nil
}

// VoiceChange registers a voice change.
// The next call to WriteNotes or WriteNote will write a voice change before the note,
// even if the voice didn't actually change compared to the previous note.
func (w *Writer) VoiceChange() {
	w.voice = -1
}

// WriteNote writes n for the specified voice.
// If the voice differs from the voice of the previous note, a voice change is inserted.
// Depending on w.Relative the note is adjusted to the current relative offset.
func (w *Writer) WriteNote(n Note, voice int) error {
	if voice < P1 || voice > P9 {
		panic("invalid voice change")
	}
	if voice != w.voice {
		for _, s := range []string{"P", strconv.Itoa(voice + 1), "\n"} {
			if _, err := io.WriteString(w.wr, s); err != nil {
				return err
			}
		}
		w.voice = voice
	}
	var parts []string
	if w.Relative {
		n.Start -= w.rel[voice]
	}
	if n.Type.IsEndOfPhrase() {
		beat := strconv.Itoa(int(n.Start))
		if w.Relative {
			parts = []string{
				string(NoteTypeEndOfPhrase),
				string(w.FieldSeparator),
				beat,
				string(w.FieldSeparator),
				beat,
				"\n",
			}
			w.rel[voice] += n.Start
		} else {
			parts = []string{
				string(NoteTypeEndOfPhrase),
				string(w.FieldSeparator),
				beat,
				"\n",
			}
		}
	} else {
		parts = []string{
			string(n.Type),
			string(w.FieldSeparator),
			strconv.Itoa(int(n.Start)),
			string(w.FieldSeparator),
			strconv.Itoa(int(n.Duration)),
			string(w.FieldSeparator),
			strconv.Itoa(int(n.Pitch)),
			string(w.FieldSeparator),
			n.Text,
			"\n",
		}
	}
	for _, p := range parts {
		if _, err := io.WriteString(w.wr, p); err != nil {
			return err
		}
	}
	return nil
}

// Close writes the final "E" line of the song.
// Anything written to w or its underlying writer after this method returns
// will be ignored by programs reading the song.
//
// This method does not close the underlying writer of w.
func (w *Writer) Close() error {
	_, err := io.WriteString(w.wr, "E\n")
	return err
}
