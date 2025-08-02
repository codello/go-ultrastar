package ultrastar

import (
	"math"
	"time"
)

// BPM is a measurement for the tempo of a song. It counts the number of [Beat]
// values per minute.
type BPM float64

// IsValid indicates whether b is a valid BPM number. Valid BPM numbers are
// finite, positive values.
func (b BPM) IsValid() bool {
	if math.IsInf(float64(b), 0) {
		return false
	}
	// This implicitly includes a NaN check
	return b > 0
}

// Beats returns the number of beats in the specified duration. The result is
// rounded down to the nearest integer. If b is invalid, the result is undefined.
func (b BPM) Beats(d time.Duration) Beat {
	return Beat(float64(b) * d.Minutes())
}

// Duration returns the time it takes for bs beats to pass. If b is invalid, the
// result is undefined.
func (b BPM) Duration(bs Beat) time.Duration {
	return time.Duration(float64(bs) / float64(b) * float64(time.Minute))
}
