package txt

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"codello.dev/ultrastar"
)

// These are the tags recognized by this package (in their canonical format).
const (
	// TagRelative is an indicator whether a song's notes must be interpreted in relative mode.
	// If this tag is absent or not set to `"YES"` the song is interpreted in absolute mode.
	TagRelative = "RELATIVE"

	// TagEncoding is a known legacy tag that specifies the encoding of a txt file.
	// UltraStar and Vocaluxe only understand the values "CP1250" and "CP1252".
	// New songs should only use UTF-8 encoding.
	TagEncoding = "ENCODING"

	// TagMP3 references the audio file for a song.
	// The value is a file path relative to the TXT file.
	// The audio file may be in MP3 format but other formats are supported as well.
	// Specifically video files may also be used.
	//
	// This tag is not required but a without this tag a song has no audio.
	TagMP3 = "MP3"

	// TagVideo references the video file for a song.
	// The value is a file path relative to the TXT file.
	TagVideo = "VIDEO"

	// TagCover references the artwork file for a song.
	// The value is a file path relative to the TXT file.
	TagCover = "COVER"

	// TagBackground references the background image for a song.
	// The value is a file path relative to the TXT file.
	TagBackground = "BACKGROUND"

	// TagBPM identifies the starting BPM for a song.
	// In most cases this BPM value holds for the entire duration of a song but
	// Multi BPM songs are supported by UltraStar.
	// The actual BPM value is 4 times as high as the value stored in the TXT file.
	//
	// The value is a floating point number.
	TagBPM = "BPM"

	// TagGap identifies the number of milliseconds before beat 0 starts.
	// This is used as an offset for the entire lyrics.
	//
	// The value is a floating point number.
	TagGap = "GAP"

	// TagVideoGap identifies the number of seconds before the video starts.
	// In contrast to TagGap this is specified in seconds instead of milliseconds.
	//
	// The value is a floating point number.
	TagVideoGap = "VIDEOGAP"

	// TagNotesGap is an obscure tag that should not be used.
	// In ultrastar this identifies an offset for the click track if you have beat clicks turned on.
	// This library treats this as a custom tag with no special meaning.
	//
	// The value is an integer.
	TagNotesGap = "NOTESGAP"

	// TagStart specifies the number of seconds into a song where singing should start.
	// This can be used for testing or to skip long intros.
	//
	// The value is a floating point number.
	TagStart = "START"

	// TagEnd specifies the number of milliseconds into a song where singing should end.
	// This can be used for testing or to skip long outros.
	//
	// The value is an integer.
	TagEnd = "END"

	// TagResolution is a tag that pops up in old documentation from time to time.
	// In TXT based songs this tag does not have any effect.
	// This tag originates from songs that were parsed from MIDI files (where the resolution does have an effect).
	// This library treats this as a custom tag with no special meaning.
	//
	// The value is an integer, an absent value is equivalent to 4.
	TagResolution = "RESOLUTION"

	// TagPreviewStart specifies the number of seconds into a song where the preview should start.
	//
	// The value is a floating point number.
	TagPreviewStart = "PREVIEWSTART"

	// TagMedleyStartBeat identifies the beat of the song where the medley starts.
	//
	// The value is an integer.
	TagMedleyStartBeat = "MEDLEYSTARTBEAT"

	// TagMedleyEndBeat identifies the beat of the song where the medley ends.
	//
	// The value is an integer.
	TagMedleyEndBeat = "MEDLEYENDBEAT"

	// TagCalcMedley can be set to "OFF" to disable the automatic medley and preview detection algorithms in UltraStar.
	// Other values are not supported.
	//
	// Manually setting medley start and end beat has the same effect.
	TagCalcMedley = "CALCMEDLEY"

	// TagTitle specifies the title/name of the song.
	TagTitle = "TITLE"

	// TagArtist specifies the artist of the song.
	TagArtist = "ARTIST"

	// TagGenre specifies the genre of the song.
	TagGenre = "GENRE"

	// TagEdition specifies the edition of the song.
	// The edition was originally intended as a way to categorize the original SingStar editions
	// but is often used as an arbitrary category value.
	TagEdition = "EDITION"

	// TagCreator identifies the creator of the song file.
	// This should be considered equal to TagAuthor.
	TagCreator = "CREATOR"

	// TagAuthor identifies the creator of the song file.
	// This should be considered equal to TagCreator.
	TagAuthor = "AUTHOR"

	// TagLanguage identifies the language of the song.
	// This does not have an impact on the song's lyrics but is only used as metadata for categorizing songs.
	TagLanguage = "LANGUAGE"

	// TagYear identifies the release year of the song.
	//
	// The value must be an integer.
	TagYear = "YEAR"

	// TagComment adds an arbitrary comment to a song.
	TagComment = "COMMENT"

	// TagDuetSingerP1 specifies the name of the first duet singer.
	// This tag should be considered equivalent to TagP1.
	TagDuetSingerP1 = "DUETSINGERP1"

	// TagDuetSingerP2 specifies the name of the second duet singer.
	// This tag should be considered equivalent to TagP2.
	TagDuetSingerP2 = "DUETSINGERP2"

	// TagP1 specifies the name of the first duet singer.
	// This tag should be considered equivalent to TagDuetSingerP1.
	TagP1 = "P1"

	// TagP2 specifies the name of the first duet singer.
	// This tag should be considered equivalent to TagDuetSingerP2.
	TagP2 = "P2"
)

// CanonicalTagName returns the normalized version of the specified tag name
// (that is: the uppercase version).
func CanonicalTagName(name string) string {
	return strings.ToUpper(name)
}

// SetTag parses the specified tag (as it would be present in an UltraStar file)
// and stores it in the appropriate field in s.
// If the tag does not correspond to any known tag it is stored in s.CustomTags.
//
// This method converts the value to appropriate data types for known values.
// If an error occurs during conversion it is returned.
// Otherwise, nil is returned.
func SetTag(s *ultrastar.Song, tag string, value string) error {
	return setTag(s, tag, value, true)
}

// setTag implements the [SetTag] function.
// This implementation allows for an additional parameter configuring whether international floats are supported.
func setTag(s *ultrastar.Song, tag string, value string, internationalFloat bool) error {
	tag = strings.ToUpper(strings.TrimSpace(tag))
	value = strings.TrimSpace(value)
	switch tag {
	case TagRelative:
		// All songs are in absolute mode. This cannot be set.
		return errors.New("read only tag: #" + TagRelative)
	case TagBPM:
		if bpm, err := parseFloat(value, internationalFloat); err != nil {
			return err
		} else {
			s.BPM = ultrastar.BPM(bpm * 4)
		}
	case TagMP3:
		s.AudioFileName = value
	case TagVideo:
		s.VideoFileName = value
	case TagCover:
		s.CoverFileName = value
	case TagBackground:
		s.BackgroundFileName = value
	case TagGap:
		if gap, err := parseFloat(value, internationalFloat); err != nil {
			return err
		} else {
			s.Gap = time.Duration(gap * float64(time.Millisecond))
		}
	case TagVideoGap:
		if videoGap, err := parseFloat(value, internationalFloat); err != nil {
			return err
		} else {
			s.VideoGap = time.Duration(videoGap * float64(time.Second))
		}
	case TagStart:
		if start, err := parseFloat(value, internationalFloat); err != nil {
			return err
		} else {
			s.Start = time.Duration(start * float64(time.Second))
		}
	case TagEnd:
		if end, err := parseFloat(value, internationalFloat); err != nil {
			return err
		} else {
			s.End = time.Duration(end * float64(time.Millisecond))
		}
	case TagPreviewStart:
		if previewStart, err := parseFloat(value, internationalFloat); err != nil {
			return err
		} else {
			s.PreviewStart = time.Duration(previewStart * float64(time.Second))
		}
	case TagMedleyStartBeat:
		if beat, err := strconv.Atoi(value); err != nil {
			return err
		} else {
			s.MedleyStartBeat = ultrastar.Beat(beat)
		}
	case TagMedleyEndBeat:
		if beat, err := strconv.Atoi(value); err != nil {
			return err
		} else {
			s.MedleyStartBeat = ultrastar.Beat(beat)
		}
	case TagCalcMedley:
		s.NoAutoMedley = strings.ToUpper(value) == "OFF"
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
		if year, err := strconv.Atoi(value); err != nil {
			return err
		} else {
			s.Year = year
		}
	case TagComment:
		s.Comment = value
	case TagP1, TagDuetSingerP1:
		s.DuetSinger1 = value
	case TagP2, TagDuetSingerP2:
		s.DuetSinger2 = value
	default:
		if s.CustomTags == nil {
			s.CustomTags = make(map[string]string)
		}
		s.CustomTags[tag] = value
	}
	return nil
}

// parseFloat converts a string from an UltraStar txt to a float. This function
// implements some special parsing behavior to parse UltraStar floats,
// specifically supporting a comma as decimal separator.
func parseFloat(s string, international bool) (float64, error) {
	if international {
		s = strings.Replace(s, ",", ".", 1)
	}
	return strconv.ParseFloat(s, 64)
}

// GetTag serializes the specified tag from song s and returns it.
// Known tags are resolved to the appropriate fields in [ultrastar.Song],
// other tags are fetched from the custom tags.
//
// This method does not differentiate between tags that are absent and tags that have an empty value.
// From an UltraStar perspective the two are identical.
// If you need to know if a custom tag exists or not, access the custom tags directly.
//
// For numeric tags an empty string is returned instead of a 0 value.
func GetTag(s ultrastar.Song, tag string) string {
	return getTag(s, tag, true)
}

// getTag implements the [GetTag] function.
// This implementation allows for an additional parameter configuring whether the comma is used as decimal separator.
func getTag(s ultrastar.Song, tag string, commaFloat bool) string {
	tag = strings.ToUpper(strings.TrimSpace(tag))
	switch tag {
	case TagRelative:
		// All songs are in absolute mode.
		return ""
	case TagBPM:
		if !s.BPM.IsValid() {
			return ""
		}
		return formatFloatTag(float64(s.BPM/4), commaFloat)
	case TagMP3:
		return s.AudioFileName
	case TagVideo:
		return s.VideoFileName
	case TagCover:
		return s.CoverFileName
	case TagBackground:
		return s.BackgroundFileName
	case TagGap:
		msec := int64(s.Gap / time.Millisecond)
		nsec := int64(s.Gap % time.Millisecond)
		v := float64(msec) + float64(nsec)/1000
		return formatFloatTag(v, commaFloat)
	case TagVideoGap:
		v := s.VideoGap.Seconds()
		return formatFloatTag(v, commaFloat)
	case TagStart:
		v := s.Start.Seconds()
		return formatFloatTag(v, commaFloat)
	case TagEnd:
		// For some reason UltraStar parses END as an integer. To preserve
		// compatibility we also serialize END as integer.
		return formatIntTag(int(s.End.Milliseconds()))
	case TagPreviewStart:
		v := s.PreviewStart.Seconds()
		return formatFloatTag(v, commaFloat)
	case TagMedleyStartBeat:
		return formatIntTag(int(s.MedleyStartBeat))
	case TagMedleyEndBeat:
		return formatIntTag(int(s.MedleyEndBeat))
	case TagCalcMedley:
		if s.NoAutoMedley {
			return "OFF"
		}
		return ""
	case TagTitle:
		return s.Title
	case TagArtist:
		return s.Artist
	case TagGenre:
		return s.Genre
	case TagEdition:
		return s.Edition
	case TagCreator, TagAuthor:
		return s.Creator
	case TagLanguage:
		return s.Language
	case TagYear:
		return formatIntTag(s.Year)
	case TagComment:
		return s.Comment
	case TagP1, TagDuetSingerP1:
		return s.DuetSinger1
	case TagP2, TagDuetSingerP2:
		return s.DuetSinger2
	default:
		return s.CustomTags[tag]
	}
}

// formatIntTag formats an integer to be used as a tag value.
// This method returns an empty string if i is 0.
func formatIntTag(i int) string {
	if i == 0 {
		return ""
	}
	return strconv.Itoa(i)
}

// formatFloatTag formats a floating point value to be used as a tag value.
// This method returns an empty string if v is 0.
func formatFloatTag(v float64, useComma bool) string {
	if v == 0 || math.IsNaN(v) || math.IsInf(v, 0) {
		return ""
	}
	s := strconv.FormatFloat(v, 'f', -1, 64)
	if useComma {
		s = strings.Replace(s, ".", ",", 1)
	}
	return s
}
