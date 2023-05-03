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
	res := s.scan()
	assert.True(t, res, "first line scan")
	assert.Equal(t, "line 1", s.text(), "first line")
	assert.Equal(t, 1, s.line(), "first line number")
	res = s.scan()
	assert.True(t, res, "second line scan")
	assert.Equal(t, "line 2", s.text(), "second line")
	assert.Equal(t, 2, s.line(), "second line number")
	s.unScan()
	assert.Equal(t, "line 1", s.text(), "first line after unscan")
	assert.Equal(t, 1, s.line(), "first line number after unscan")
	res = s.scan()
	assert.True(t, res, "second line scan after unscan")
	assert.Equal(t, "line 2", s.text(), "second line after unscan")
	assert.Equal(t, 2, s.line(), "second line number after unscan")
	res = s.scan()
	assert.True(t, res, "third line scan")
	assert.Equal(t, "line 3", s.text(), "third line")
	assert.Equal(t, 3, s.line(), "third line number")
	res = s.scan()
	assert.False(t, res, "overscan")
	assert.Equal(t, "", s.text(), "third line overscan")
	assert.Equal(t, 4, s.line(), "third line number overscan")
}
