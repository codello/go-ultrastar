package ultrastar

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNoteType_IsValid(t *testing.T) {
	cases := []struct {
		name     string
		nType    NoteType
		expected bool
	}{
		{"regular note", ':', true},
		{"golden note", '*', true},
		{"rap note", 'R', true},
		{"golden rap note", 'G', true},
		{"freestyle note", 'F', true},
		{"line break", '-', true},
		{"letter X", 'X', false},
		{"letter A", 'A', false},
		{"space", ' ', false},
		{"number 1", '1', false},
		{"dot", '.', false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.expected {
				assert.True(t, c.nType.IsValid())
			} else {
				assert.False(t, c.nType.IsValid())
			}
		})
	}
}

func TestNoteType_IsSung(t *testing.T) {
	cases := []struct {
		name     string
		nType    NoteType
		expected bool
		panic    bool
	}{
		{"regular note", NoteTypeRegular, true, false},
		{"golden note", NoteTypeGolden, true, false},
		{"rap note", NoteTypeRap, false, false},
		{"golden rap note", NoteTypeGoldenRap, false, false},
		{"freestyle note", NoteTypeFreestyle, false, false},
		{"line break", NoteTypeLineBreak, false, false},
		{"invalid note", '#', false, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.panic {
				assert.Panics(t, func() {
					_ = c.nType.IsSung()
				})
			} else if c.expected {
				assert.True(t, c.nType.IsSung())
			} else {
				assert.False(t, c.nType.IsSung())
			}
		})
	}
}

func TestNoteType_IsRap(t *testing.T) {
	cases := []struct {
		name     string
		nType    NoteType
		expected bool
		panic    bool
	}{
		{"regular note", NoteTypeRegular, false, false},
		{"golden note", NoteTypeGolden, false, false},
		{"rap note", NoteTypeRap, true, false},
		{"golden rap note", NoteTypeGoldenRap, true, false},
		{"freestyle note", NoteTypeFreestyle, false, false},
		{"line break", NoteTypeLineBreak, false, false},
		{"invalid note", '#', false, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.panic {
				assert.Panics(t, func() {
					_ = c.nType.IsRap()
				})
			} else if c.expected {
				assert.True(t, c.nType.IsRap())
			} else {
				assert.False(t, c.nType.IsRap())
			}
		})
	}
}

func TestNoteType_IsGolden(t *testing.T) {
	cases := []struct {
		name     string
		nType    NoteType
		expected bool
		panic    bool
	}{
		{"regular note", NoteTypeRegular, false, false},
		{"golden note", NoteTypeGolden, true, false},
		{"rap note", NoteTypeRap, false, false},
		{"golden rap note", NoteTypeGoldenRap, true, false},
		{"freestyle note", NoteTypeFreestyle, false, false},
		{"line break", NoteTypeLineBreak, false, false},
		{"invalid note", '#', false, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.panic {
				assert.Panics(t, func() {
					_ = c.nType.IsGolden()
				})
			} else if c.expected {
				assert.True(t, c.nType.IsGolden())
			} else {
				assert.False(t, c.nType.IsGolden())
			}
		})
	}
}

func TestNoteType_IsFreestyle(t *testing.T) {
	cases := []struct {
		name     string
		nType    NoteType
		expected bool
		panic    bool
	}{
		{"regular note", NoteTypeRegular, false, false},
		{"golden note", NoteTypeGolden, false, false},
		{"rap note", NoteTypeRap, false, false},
		{"golden rap note", NoteTypeGoldenRap, false, false},
		{"freestyle note", NoteTypeFreestyle, true, false},
		{"line break", NoteTypeLineBreak, false, false},
		{"invalid note", '#', false, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.panic {
				assert.Panics(t, func() {
					_ = c.nType.IsFreestyle()
				})
			} else if c.expected {
				assert.True(t, c.nType.IsFreestyle())
			} else {
				assert.False(t, c.nType.IsFreestyle())
			}
		})
	}
}

func TestNoteType_IsLineBreak(t *testing.T) {
	cases := []struct {
		name     string
		nType    NoteType
		expected bool
		panic    bool
	}{
		{"regular note", NoteTypeRegular, false, false},
		{"golden note", NoteTypeGolden, false, false},
		{"rap note", NoteTypeRap, false, false},
		{"golden rap note", NoteTypeGoldenRap, false, false},
		{"freestyle note", NoteTypeFreestyle, false, false},
		{"line break", NoteTypeLineBreak, true, false},
		{"invalid note", '#', false, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.panic {
				assert.Panics(t, func() {
					_ = c.nType.IsLineBreak()
				})
			} else if c.expected {
				assert.True(t, c.nType.IsLineBreak())
			} else {
				assert.False(t, c.nType.IsLineBreak())
			}
		})
	}
}

func TestNote_String(t *testing.T) {
	cases := []struct {
		name     string
		note     Note
		expected string
	}{
		{"regular note", Note{NoteTypeRegular, 15, 4, 8, "go"}, ": 15 4 8 go"},
		{"note with spaces in text", Note{NoteTypeGolden, 7, 1, -2, " hey "}, "* 7 1 -2  hey "},
		{"line break", Note{NoteTypeLineBreak, 12, 7, 3, "\n"}, "- 12"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, c.note.String())
		})
	}
}
