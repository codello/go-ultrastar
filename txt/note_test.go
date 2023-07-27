package txt

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Karaoke-Manager/go-ultrastar"
)

func TestParseNote(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ultrastar.Note
		wantErr assert.ErrorAssertionFunc
	}{
		{"regular note", ": 5 2 3 some", ultrastar.Note{Type: ultrastar.NoteTypeRegular, Start: 5, Duration: 2, Pitch: 3, Text: "some"}, assert.NoError},
		{"leading space", "* 6 2 1  body", ultrastar.Note{Type: ultrastar.NoteTypeGolden, Start: 6, Duration: 2, Pitch: 1, Text: " body"}, assert.NoError},
		{"trailing space", "R 2 11 6 once  ", ultrastar.Note{Type: ultrastar.NoteTypeRap, Start: 2, Duration: 11, Pitch: 6, Text: "once  "}, assert.NoError},
		{"multiple spaces", "G  1    2    52    told ", ultrastar.Note{Type: ultrastar.NoteTypeGoldenRap, Start: 1, Duration: 2, Pitch: 52, Text: "   told "}, assert.NoError},
		{"no space after note type", "*12 41 3 me, ", ultrastar.Note{Type: ultrastar.NoteTypeGolden, Start: 12, Duration: 41, Pitch: 3, Text: "me, "}, assert.NoError},
		{"missing note type", "", ultrastar.Note{}, assert.Error},
		{"missing note start", ":", ultrastar.Note{Type: ultrastar.NoteTypeRegular}, assert.Error},
		{"missing note duration", ": 12", ultrastar.Note{Type: ultrastar.NoteTypeRegular, Start: 12}, assert.Error},
		{"missing note pitch", ": 12 3", ultrastar.Note{Type: ultrastar.NoteTypeRegular, Start: 12, Duration: 3}, assert.Error},
		{"missing note text", ": 12 3 6", ultrastar.Note{Type: ultrastar.NoteTypeRegular, Start: 12, Duration: 3, Pitch: 6}, assert.Error},
		{"float note start", ": 23.4 1 3 Hello", ultrastar.Note{Type: ultrastar.NoteTypeRegular}, assert.Error},
		{"invalid note type", "X 3 5 1 World", ultrastar.Note{}, assert.Error},
		{"missing space", ": 5 4 3test", ultrastar.Note{Type: ultrastar.NoteTypeRegular, Start: 5, Duration: 4}, assert.Error},
		{"line break", "- 52", ultrastar.Note{Type: ultrastar.NoteTypeLineBreak, Start: 52, Text: "\n"}, assert.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseNote(tt.input)
			if !tt.wantErr(t, err, fmt.Sprintf("ParseNote(%#v)", tt.input)) {
				return
			}
			assert.Equalf(t, tt.want, got, "ParseNote(%#v)", tt.input)
		})
	}
}
