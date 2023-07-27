package txt

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding/charmap"
)

func TestTransformSong(t *testing.T) {
	f, _ := os.Open("testdata/Juli - Perfekte Welle.txt")
	defer f.Close()
	d := new(Dialect)
	*d = *DialectDefault
	d.ApplyEncoding = false
	s, err := d.ReadSong(f)
	require.NoError(t, err)

	err = TransformSong(s, charmap.Windows1252.NewDecoder())
	assert.NoError(t, err)
	assert.Equal(t, " Träu", s.MusicP1.Notes[10].Text)
}
