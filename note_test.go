package ultrastar

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"testing"
)

/*
func TestNoteType_IsStandard(t *testing.T) {
	cases := map[string]struct {
		nType    NoteType
		expected bool
	}{
		"regular note":    {':', true},
		"golden note":     {'*', true},
		"rap note":        {'R', true},
		"golden rap note": {'G', true},
		"freestyle note":  {'F', true},
		"line break":      {'-', true},
		"letter X":        {'X', false},
		"letter A":        {'A', false},
		"space":           {' ', false},
		"number 1":        {'1', false},
		"dot":             {'.', false},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			actual := c.nType.IsStandard()
			if actual != c.expected {
				t.Errorf("NoteType('%c').IsStandard() = %t, expected %t", c.nType, actual, c.expected)
			}
		})
	}
}

func TestNoteType_IsSung(t *testing.T) {
	cases := map[string]struct {
		nType    NoteType
		expected bool
		panic    bool
	}{
		"regular note":    {NoteTypeRegular, true, false},
		"golden note":     {NoteTypeGolden, true, false},
		"rap note":        {NoteTypeRap, false, false},
		"golden rap note": {NoteTypeGoldenRap, false, false},
		"freestyle note":  {NoteTypeFreestyle, false, false},
		"line break":      {NoteTypeEndOfPhrase, false, false},
		"invalid note":    {'#', false, true},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil && c.panic {
					t.Errorf("NoteType('%c').IsSung() did not panic", c.nType)
				} else if r != nil && !c.panic {
					t.Errorf("NoteType('%c').IsSung() caused a panic", c.nType)
				}
			}()
			actual := c.nType.IsSung()
			if actual != c.expected {
				t.Errorf("NoteType('%c').IsSung() = %t, expected %t", c.nType, actual, c.expected)
			}
		})
	}
}

func TestNoteType_IsRap(t *testing.T) {
	cases := map[string]struct {
		nType    NoteType
		expected bool
		panic    bool
	}{
		"regular note":    {NoteTypeRegular, false, false},
		"golden note":     {NoteTypeGolden, false, false},
		"rap note":        {NoteTypeRap, true, false},
		"golden rap note": {NoteTypeGoldenRap, true, false},
		"freestyle note":  {NoteTypeFreestyle, false, false},
		"line break":      {NoteTypeEndOfPhrase, false, false},
		"invalid note":    {'#', false, true},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil && c.panic {
					t.Errorf("NoteType('%c').IsRap() did not panic", c.nType)
				} else if r != nil && !c.panic {
					t.Errorf("NoteType('%c').IsRap() caused a panic", c.nType)
				}
			}()
			actual := c.nType.IsRap()
			if actual != c.expected {
				t.Errorf("NoteType('%c').IsRap() = %t, expected %t", c.nType, actual, c.expected)
			}
		})
	}
}

func TestNoteType_IsGolden(t *testing.T) {
	cases := map[string]struct {
		nType    NoteType
		expected bool
		panic    bool
	}{
		"regular note":    {NoteTypeRegular, false, false},
		"golden note":     {NoteTypeGolden, true, false},
		"rap note":        {NoteTypeRap, false, false},
		"golden rap note": {NoteTypeGoldenRap, true, false},
		"freestyle note":  {NoteTypeFreestyle, false, false},
		"line break":      {NoteTypeEndOfPhrase, false, false},
		"invalid note":    {'#', false, true},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil && c.panic {
					t.Errorf("NoteType('%c').IsGolden() did not panic", c.nType)
				} else if r != nil && !c.panic {
					t.Errorf("NoteType('%c').IsGolden() caused a panic", c.nType)
				}
			}()
			actual := c.nType.IsGolden()
			if actual != c.expected {
				t.Errorf("NoteType('%c').IsGolden() = %t, expected %t", c.nType, actual, c.expected)
			}
		})
	}
}

func TestNoteType_IsFreestyle(t *testing.T) {
	cases := map[string]struct {
		nType    NoteType
		expected bool
		panic    bool
	}{
		"regular note":    {NoteTypeRegular, false, false},
		"golden note":     {NoteTypeGolden, false, false},
		"rap note":        {NoteTypeRap, false, false},
		"golden rap note": {NoteTypeGoldenRap, false, false},
		"freestyle note":  {NoteTypeFreestyle, true, false},
		"line break":      {NoteTypeEndOfPhrase, false, false},
		"invalid note":    {'#', false, true},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil && c.panic {
					t.Errorf("NoteType('%c').IsFreestyle() did not panic", c.nType)
				} else if r != nil && !c.panic {
					t.Errorf("NoteType('%c').IsFreestyle() caused a panic", c.nType)
				}
			}()
			actual := c.nType.IsFreestyle()
			if actual != c.expected {
				t.Errorf("NoteType('%c').IsFreestyle() = %t, expected %t", c.nType, actual, c.expected)
			}
		})
	}
}

func TestNoteType_IsLineBreak(t *testing.T) {
	cases := map[string]struct {
		nType    NoteType
		expected bool
		panic    bool
	}{
		"regular note":    {NoteTypeRegular, false, false},
		"golden note":     {NoteTypeGolden, false, false},
		"rap note":        {NoteTypeRap, false, false},
		"golden rap note": {NoteTypeGoldenRap, false, false},
		"freestyle note":  {NoteTypeFreestyle, false, false},
		"line break":      {NoteTypeEndOfPhrase, true, false},
		"invalid note":    {'#', false, true},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil && c.panic {
					t.Errorf("NoteType('%c').IsEndOfPhrase() did not panic", c.nType)
				} else if r != nil && !c.panic {
					t.Errorf("NoteType('%c').IsEndOfPhrase() caused a panic", c.nType)
				}
			}()
			actual := c.nType.IsEndOfPhrase()
			if actual != c.expected {
				t.Errorf("NoteType('%c').IsEndOfPhrase() = %t, expected %t", c.nType, actual, c.expected)
			}
		})
	}
}*/

func TestNote_String(t *testing.T) {
	cases := map[string]struct {
		note     Note
		expected string
	}{
		"regular note":             {Note{NoteTypeRegular, 15, 4, 8, "go"}, ": 15 4 8 go"},
		"note with spaces in text": {Note{NoteTypeGolden, 7, 1, -2, " hey "}, "* 7 1 -2  hey "},
		"line break":               {Note{NoteTypeEndOfPhrase, 12, 7, 3, "\n"}, "- 12"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			actual := c.note.String()
			if actual != c.expected {
				t.Errorf("%v.String() = %q, expected %q", c.note, actual, c.expected)
			}
		})
	}
}

func TestNote_GobEncode(t *testing.T) {
	cases := map[string]Note{
		"regular note":             Note{NoteTypeRegular, 15, 4, 8, "go"},
		"note with spaces in text": Note{NoteTypeGolden, 7, 1, -2, " hey "},
		"note with high numbers":   Note{NoteTypeRap, 550, 20, -40, " hey "},
		"line break":               Note{NoteTypeEndOfPhrase, 12, 0, 0, "\n"},
	}
	for name, note := range cases {
		t.Run(name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			e := gob.NewEncoder(buf)
			err := e.Encode(note)
			if err != nil {
				t.Fatalf("GobEncode(%v) caused an unexpected error: %s", note, err)
			}

			var n Note
			d := gob.NewDecoder(buf)
			err = d.Decode(&n)
			if err != nil {
				t.Fatalf("GobDecode() caused an unexpected error: %s", err)
			}
			if n != note {
				t.Fatalf("GobDecode(GobEncode(%v)) = %v, expected %v", note, n, note)
			}
		})
	}
}

func ExampleNote_String() {
	n := Note{
		Type:     NoteTypeGolden,
		Start:    15,
		Duration: 4,
		Pitch:    8,
		Text:     "Go",
	}
	fmt.Println(n.String())
	// Output: * 15 4 8 Go
}
