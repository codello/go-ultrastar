package txt

import (
	"github.com/codello/ultrastar"
	"strconv"
	"strings"
	"time"
)

const (
	// TagRelative is an indicator whether a song's music must be interpreted in
	// relative mode. If this tag is absent or not set to `"YES"` the song is
	// interpreted in absolute mode.
	TagRelative = "RELATIVE"

	// TagEncoding is a known legacy tag that specifies the encoding of a txt
	// file. UltraStar and Vocaluxe only understand the values `"CP1250"` and
	// `"CP1252"`
	TagEncoding = "ENCODING"

	// TagMP3 references the audio file for a song. The value is a file path
	// relative to the TXT file. The audio file may be in MP3 format but other
	// formats are supported as well. Specifically video files may also be used.
	//
	// This tag is not required but a without this tag a song has no audio.
	TagMP3 = "MP3"

	// TagVideo references the video file for a song. The value is a file path
	// relative to the TXT file.
	TagVideo = "VIDEO"

	// TagCover references the artwork file for a song. The value is a file path
	// relative to the TXT file.
	TagCover = "COVER"

	// TagBackground references the background image for a song. The value is a
	// file path relative to the TXT file.
	TagBackground = "BACKGROUND"

	// TagBPM identifies the starting BPM for a song. In most cases this BPM
	// value holds for the entire duration of a song but Multi BPM songs are
	// supported by UltraStar. The actual BPM value is 4 times as high as the
	// value stored in the TXT file.
	//
	// The value is a floating point number.
	TagBPM = "BPM"

	// TagGap identifies the number of milliseconds before beat 0 starts. This
	// is used as an offset for the entire lyrics.
	//
	// The value is a floating point number.
	TagGap = "GAP"

	// TagVideoGap identifies the number of seconds before the video starts. In
	// contrast to [TagGap] this is specified in seconds instead of milliseconds.
	//
	// The value is a floating point number.
	TagVideoGap = "VIDEOGAP"

	// TagNotesGap identifies some kind of Beat offset for notes. The exact
	// purpose is currently unclear.
	//
	// The value is an integer.
	TagNotesGap = "NOTESGAP"

	// TagStart specifies the number of seconds into a song where singing should
	// start. This can be used for testing or to skip long intros.
	//
	// The value is a floating point number.
	TagStart = "START"

	// TagEnd specifies the number of milliseconds into a song where singing
	// should end. This can be used for testing or to skip long outros.
	//
	// The value is an integer.
	TagEnd = "END"

	// TagResolution seems to be relevant only in XML formatted songs. The exact
	// purpose is unclear.
	//
	// The value is an integer.
	TagResolution = "RESOLUTION"

	// TagPreviewStart specifies the number of seconds into a song where the
	// preview should start.
	//
	// The value is a floating point number.
	TagPreviewStart = "PREVIEWSTART"

	// TagMedleyStartBeat identifies the beat of the song where the medley
	// starts.
	//
	// The value is an integer.
	TagMedleyStartBeat = "MEDLEYSTARTBEAT"

	// TagMedleyEndBeat identifies the beat of the song where the medley ends.
	//
	// The value is an integer.
	TagMedleyEndBeat = "MEDLEYENDBEAT"

	// TagCalcMedley can be set to `"OFF"` to disable the automatic medley and
	// preview detection algorithms in UltraStar. Other values are not supported.
	//
	// Manually setting medley start and end beat has the same effect.
	TagCalcMedley = "CALCMEDLEY"

	// TagTitle specifies the title/name of the song.
	TagTitle = "TITLE"

	// TagArtist specifies the artist of the song.
	TagArtist = "ARTIST"

	// TagGenre specifies the genre of the song.
	TagGenre = "GENRE"

	// TagEdition specifies the edition of the song. The edition was originally
	// intended as a way to categorize the original SingStar editions but can
	// be used as an arbitrary category value.
	TagEdition = "EDITION"

	// TagCreator identifies the creator of the song file. This should be
	// considered equal to [TagAuthor].
	TagCreator = "CREATOR"

	// TagAuthor identifies the creator of the song file. This should be
	// considered equal to [TagCreator].
	TagAuthor = "AUTHOR"

	// TagLanguage identifies the language of the song. This does not have an
	// impact on the song's lyrics but is only used as metadata for categorizing
	// songs.
	TagLanguage = "LANGUAGE"

	// TagYear identifies the release year of the song.
	//
	// The value must be an integer.
	TagYear = "YEAR"

	// TagComment adds an arbitrary comment to a song.
	TagComment = "COMMENT"

	// TagDuetSingerP1 specifies the name of the first duet singer. This tag
	// should be considered equivalent to [TagP1].
	TagDuetSingerP1 = "DUETSINGERP1"

	// TagDuetSingerP2 specifies the name of the second duet singer. This tag
	// should be considered equivalent to [TagP2].
	TagDuetSingerP2 = "DUETSINGERP2"

	// TagP1 specifies the name of the first duet singer. This tag should be
	// considered equivalent to [TagDuetSingerP1].
	TagP1 = "P1"

	// TagP2 specifies the name of the first duet singer. This tag should be
	// considered equivalent to [TagDuetSingerP2].
	TagP2 = "P2"
)

// SetTag parses the specified tag (as it would be present in an UltraStar file)
// and stores it in the appropriate field in s. If the tag does not correspond
// to any known tag it is stored in [ultrastar.Song.CustomTags] of s.
//
// This method converts the value to appropriate data types for known values. If
// an error occurs during conversion it is returned. Otherwise nil is returned.
func SetTag(s *ultrastar.Song, tag string, value string) error {
	tag = strings.ToUpper(strings.TrimSpace(tag))
	value = strings.TrimSpace(value)
	switch tag {
	case TagRelative, TagEncoding, TagBPM:
		// These tags are processed by the parser and ignored here
	case TagMP3:
		s.AudioFile = value
	case TagVideo:
		s.VideoFile = value
	case TagCover:
		s.CoverFile = value
	case TagBackground:
		s.BackgroundFile = value
	case TagGap:
		gap, err := parseFloat(value)
		s.Gap = time.Duration(gap * float64(time.Millisecond))
		return err
	case TagVideoGap:
		videoGap, err := parseFloat(value)
		s.VideoGap = time.Duration(videoGap * float64(time.Second))
		return err
	case TagNotesGap:
		notesGap, err := strconv.Atoi(value)
		s.NotesGap = ultrastar.Beat(notesGap)
		return err
	case TagStart:
		start, err := parseFloat(value)
		s.Start = time.Duration(start * float64(time.Second))
		return err
	case TagEnd:
		end, err := parseFloat(value)
		s.End = time.Duration(end * float64(time.Millisecond))
		return err
	case TagPreviewStart:
		previewStart, err := parseFloat(value)
		s.PreviewStart = time.Duration(previewStart * float64(time.Second))
		return err
	case TagMedleyStartBeat:
		beat, err := strconv.Atoi(value)
		s.MedleyStartBeat = ultrastar.Beat(beat)
		return err
	case TagMedleyEndBeat:
		beat, err := strconv.Atoi(value)
		s.MedleyStartBeat = ultrastar.Beat(beat)
		return err
	case TagResolution:
		res, err := strconv.Atoi(value)
		s.Resolution = res
		return err
	case TagCalcMedley:
		s.CalcMedley = strings.ToUpper(value) != "OFF"
	case TagTitle:
		s.Title = value
	case TagArtist:
		s.Artist = value
	case TagGenre:
		s.Genre = value
	case TagEdition:
		s.Edition = value
	case TagCreator, TagAuthor:
		s.Creator = value
	case TagLanguage:
		s.Language = value
	case TagYear:
		year, err := strconv.Atoi(value)
		s.Year = year
		return err
	case TagComment:
		s.Comment = value
	case TagP1, TagDuetSingerP1:
		s.DuetSinger1 = value
	case TagP2, TagDuetSingerP2:
		s.DuetSinger2 = value
	default:
		s.CustomTags[tag] = value
	}
	return nil
}
