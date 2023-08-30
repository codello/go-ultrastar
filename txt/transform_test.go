package txt

import (
	"os"
	"testing"

	"golang.org/x/text/encoding/charmap"
)

func TestTransformSong(t *testing.T) {
	f, _ := os.Open("testdata/Juli - Perfekte Welle.txt")
	defer f.Close()
	d := DialectDefault
	d.ApplyEncoding = false
	s, _ := d.ReadSong(f)

	err := TransformSong(s, charmap.Windows1252.NewDecoder())
	if err != nil {
		t.Errorf("TransformSong(s, \"CP1252\") caused an unexpected error: %s", err)
	}
	if s.NotesP1[10].Text != " Träu" {
		t.Errorf("TransformSong(s, \"CP1252\") produced %q, expected %q", s.NotesP1[10].Text, " Träu")
	}
}
