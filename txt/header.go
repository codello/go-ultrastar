package txt

import (
	"errors"
	"github.com/codello/ultrastar"
	"strconv"
	"strings"
	"time"
)

// ParseFloat converts a string from an UltraStar txt to a float. This function
// implements some special parsing behavior to parse UltraStar floats,
// specifically supporting a comma as decimal separator.
func ParseFloat(s string) (float64, error) {
	return strconv.ParseFloat(strings.Replace(s, ",", ".", 1), 64)
}

type Header map[string]string

var (
	ErrTagNotPresent = errors.New("tag not present")
)

func (h Header) Get(tag string) string {
	return h[CanonicalTagName(tag)]
}

func (h Header) String(tag string) (string, error) {
	if value, ok := h[CanonicalTagName(tag)]; ok {
		return value, nil
	} else {
		return "", ErrTagNotPresent
	}
}

func (h Header) Int(tag string) (int, error) {
	value := h.Get(tag)
	if value == "" {
		return 0, ErrTagNotPresent
	}
	return strconv.Atoi(value)
}

func (h Header) Float(tag string) (float64, error) {
	value := h.Get(tag)
	if value == "" {
		return 0, ErrTagNotPresent
	}
	return ParseFloat(value)
}

func (h Header) Beat(tag string) (ultrastar.Beat, error) {
	value, err := h.Int(tag)
	return ultrastar.Beat(value), err
}

func (h Header) Seconds(tag string) (time.Duration, error) {
	value, err := h.Float(tag)
	return time.Duration(value * float64(time.Second)), err
}

func (h Header) Milliseconds(tag string) (time.Duration, error) {
	value, err := h.Float(tag)
	return time.Duration(value * float64(time.Millisecond)), err
}

func (h Header) Set(tag string, value string) {
	if value == "" {
		h.Del(tag)
	} else {
		h[CanonicalTagName(tag)] = value
	}
}

func (h Header) SetInt(tag string, value int) {
	h.Set(tag, strconv.Itoa(value))
}

func (h Header) SetFloat(tag string, value float64) {
	h.Set(tag, strconv.FormatFloat(value, 'f', -1, 64))
}

func (h Header) SetBeat(tag string, value ultrastar.Beat) {
	h.SetInt(tag, int(value))
}

func (h Header) SetSeconds(tag string, value time.Duration) {
	h.SetFloat(tag, value.Seconds())
}

func (h Header) SetMilliseconds(tag string, value time.Duration) {
	h.SetInt(tag, int(value.Milliseconds()))
}

func (h Header) Del(tag string) {
	delete(h, CanonicalTagName(tag))
}
