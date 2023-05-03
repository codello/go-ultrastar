package ultrastar

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNoteType_IsValid_Positive(t *testing.T) {
	tests := []NoteType{':', '*', 'R', 'G', 'F'}
	for _, test := range tests {
		assert.True(t, test.IsValid(), string(test))
	}
}

func TestNoteType_IsValid_Negative(t *testing.T) {
	tests := []NoteType{'X', 'A', ' ', '1', '.'}
	for _, test := range tests {
		assert.False(t, test.IsValid(), string(test))
	}
}

func TestNoteType_IsSung(t *testing.T) {
	assert.True(t, NoteTypeRegular.IsSung(), "regular note")
	assert.True(t, NoteTypeGolden.IsSung(), "golden note")
	assert.False(t, NoteTypeRap.IsSung(), "rap note")
	assert.False(t, NoteTypeGoldenRap.IsSung(), "golden rap note")
	assert.False(t, NoteTypeFreestyle.IsSung(), "freestyle note")
}

func TestNoteType_IsRap(t *testing.T) {
	assert.True(t, NoteTypeRap.IsRap(), "rap note")
	assert.True(t, NoteTypeGoldenRap.IsRap(), "golden rap note")
	assert.False(t, NoteTypeRegular.IsRap(), "regular note")
	assert.False(t, NoteTypeGolden.IsRap(), "golden note")
	assert.False(t, NoteTypeFreestyle.IsRap(), "freestyle note")
}

func TestNoteType_IsGolden(t *testing.T) {
	assert.True(t, NoteTypeGolden.IsGolden(), "golden note")
	assert.True(t, NoteTypeGoldenRap.IsGolden(), "golden rap note")
	assert.False(t, NoteTypeRegular.IsGolden(), "regular note")
	assert.False(t, NoteTypeRap.IsGolden(), "rap note")
	assert.False(t, NoteTypeFreestyle.IsGolden(), "freestyle note")
}

func TestNoteType_IsFreestyle(t *testing.T) {
	assert.True(t, NoteTypeFreestyle.IsFreestyle(), "freestyle note")
	assert.False(t, NoteTypeRegular.IsFreestyle(), "regular note")
	assert.False(t, NoteTypeGolden.IsFreestyle(), "golden note")
	assert.False(t, NoteTypeRap.IsFreestyle(), "rap note")
	assert.False(t, NoteTypeGoldenRap.IsFreestyle(), "golden rap note")
}

func TestNote_String(t *testing.T) {
	n := Note{
		Type:     NoteTypeRegular,
		Start:    15,
		Duration: 4,
		Pitch:    NamedPitch("G#4"),
		Text:     " hey ",
	}
	assert.Equal(t, ": 15 4 8  hey ", n.String())
}
