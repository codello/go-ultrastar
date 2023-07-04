package txt

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestScanner(t *testing.T) {
	s := newScanner(strings.NewReader(`line 1
line 2
line 3`))
	res := s.Scan()
	assert.True(t, res, "first line scan")
	assert.Equal(t, "line 1", s.Text(), "first line")
	assert.Equal(t, 1, s.Line(), "first line number")
	res = s.Scan()
	assert.True(t, res, "second line scan")
	assert.Equal(t, "line 2", s.Text(), "second line")
	assert.Equal(t, 2, s.Line(), "second line number")
	s.UnScan()
	assert.Equal(t, "line 1", s.Text(), "first line after unscan")
	assert.Equal(t, 1, s.Line(), "first line number after unscan")
	res = s.Scan()
	assert.True(t, res, "second line scan after unscan")
	assert.Equal(t, "line 2", s.Text(), "second line after unscan")
	assert.Equal(t, 2, s.Line(), "second line number after unscan")
	res = s.Scan()
	assert.True(t, res, "third line scan")
	assert.Equal(t, "line 3", s.Text(), "third line")
	assert.Equal(t, 3, s.Line(), "third line number")
	res = s.Scan()
	assert.False(t, res, "overscan")
	assert.Equal(t, "", s.Text(), "third line overscan")
	assert.Equal(t, 4, s.Line(), "third line number overscan")
}
