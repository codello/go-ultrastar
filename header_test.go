package ultrastar

import (
	"errors"
	"slices"
	"strconv"
	"testing"
)

func TestCanonicalHeaderKey(t *testing.T) {
	tests := []struct{ key, expected string }{
		{"TITLE", "TITLE"},
		{"artist", "ARTIST"},
		{"gEnRe", "GENRE"},
		{"FOO-bar", "FOO-BAR"},
		{"inval:d", ""},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			actual := CanonicalHeaderKey(tt.key)
			if actual != tt.expected {
				t.Errorf("CanonicalHeaderKey(%q) = %q, expected %q", tt.key, actual, tt.expected)
			}
		})
	}
}

func TestHeader_Add(t *testing.T) {
	h := make(Header)
	h.Add("title", "hello world")
	if len(h[HeaderTitle]) != 1 || h[HeaderTitle][0] != "hello world" {
		t.Errorf("h.Add(%q, %q) did not append the expected value", "title", "hello world")
	}
}

func TestHeader_Set(t *testing.T) {
	h := Header{
		HeaderGenre: []string{"Rock", "Pop"},
	}
	h.Set(HeaderGenre, "Soul")
	if vs, ok := h[HeaderGenre]; !ok || len(vs) != 1 || vs[0] != "Soul" {
		t.Errorf("h.Set(%q, %q) did not replace the previous value", HeaderGenre, "Soul")
	}
}

func TestHeader_SetValues(t *testing.T) {
	h := make(Header)
	h.SetValues("test", []string{"foo", "", "bar"})
	if len(h["TEST"]) != 2 {
		t.Errorf("h.SetValues(...) did not remove empty values, expected no empty values")
	}
}

func TestHeader_Get(t *testing.T) {
	h := Header{
		HeaderAudio: []string{"Hello", "World"},
	}
	if h.Get("audio") == "" {
		t.Errorf("h.Get(...) = \"\", expected non-empty string")
	}
}

func TestHeader_Del(t *testing.T) {
	h := Header{
		HeaderArtist: []string{"Hello", "World"},
	}
	h.Del("foobar")
	if len(h) != 1 {
		t.Errorf("h.Del(...) deleted an element when it should not")
	}
	h.Del("artist")
	if len(h) != 0 {
		t.Errorf("h.Del(...) did not delete an element when it should have")
	}
}

func TestHeader_Len(t *testing.T) {
	h := Header{
		HeaderArtist: []string{"", "Hello", ""},
		"artist":     []string{"World"},
		"foo:bar":    []string{"removed"},
		"":           []string{"removed"},
	}
	h.Clean()
	if len(h) != 1 {
		t.Errorf("h.Clean() left %d elements, expected 1", len(h))
	}
	if len(h[HeaderArtist]) != 2 {
		t.Errorf("h.Clean() left %v, expected [Hello World]", h[HeaderArtist])
	}
}

func TestHeader_Unique(t *testing.T) {
	tests := []struct {
		expected string
		err      error
		values   []string
	}{
		{"foo", nil, []string{"foo"}},
		{"bar", nil, []string{"bar", "bar", ""}},
		{"foo", nil, []string{"", "foo", ""}},
		{"", nil, []string{}},
		{"", nil, []string{"", ""}},
		{"", ErrMultipleValues, []string{"foo", "bar"}},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			h := Header{
				HeaderTitle: tt.values,
			}
			actual, err := h.GetUnique(HeaderTitle)
			if tt.err != nil && !errors.Is(err, tt.err) {
				t.Errorf("h.GetUnique(...) = _, %q, expected %q", err, tt.err)
			} else if tt.err == nil && actual != tt.expected {
				t.Errorf("h.GetUnique(...) = %q, _, expected %q", actual, tt.expected)
			}
		})
	}
}

func TestHeader_GetMultiValued(t *testing.T) {
	tests := []struct {
		raw      []string
		expected []string
	}{
		{[]string{"foo,bar", "foobar"}, []string{"foo", "bar", "foobar"}},
		{[]string{"foo,,,bar,"}, []string{"foo,", "bar"}},
		{[]string{"  bar  ,  foo ,   , ,", ",foo,"}, []string{"bar", "foo", "foo"}},
		{[]string{"Foo", ""}, []string{"Foo"}},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			h := Header{
				HeaderGenre: tt.raw,
			}
			actual := slices.Collect(h.GetMultiValued(HeaderGenre))
			if !slices.Equal(actual, tt.expected) {
				t.Errorf("h.GetMultiValued(...) = %v, expected %v", actual, tt.expected)
			}
		})
	}
}
