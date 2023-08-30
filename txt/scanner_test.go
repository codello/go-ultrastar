package txt

import (
	"strings"
	"testing"
)

func TestScanner(t *testing.T) {
	s := newScanner(strings.NewReader(`line 1
line 2
line 3`))

	res := s.Scan()
	line, no := s.Text(), s.Line()
	if !res {
		t.Errorf("line 1: s.Scan() = false, expected true")
	}
	if line != "line 1" {
		t.Errorf("line 1: s.Text() = %q, expected %q", line, "line 1")
	}
	if no != 1 {
		t.Errorf("line 1: s.Line() = %d, expected %d", no, 1)
	}

	res = s.Scan()
	line, no = s.Text(), s.Line()
	if !res {
		t.Errorf("line 2: s.Scan() = false, expected true")
	}
	if line != "line 2" {
		t.Errorf("line 2: s.Text() = %q, expected %q", line, "line 2")
	}
	if no != 2 {
		t.Errorf("line 2: s.Line() = %d, expected %d", no, 2)
	}

	s.UnScan()
	line, no = s.Text(), s.Line()
	if line != "line 1" {
		t.Errorf("line 1 (unscan): s.Text() = %q, expected %q", line, "line 1")
	}
	if no != 1 {
		t.Errorf("line 1 (unscan): s.Line() = %d, expected %d", no, 1)
	}

	res = s.Scan()
	line, no = s.Text(), s.Line()
	if !res {
		t.Errorf("line 2 (rescan): s.Scan() = false, expected true")
	}
	if line != "line 2" {
		t.Errorf("line 2 (rescan): s.Text() = %q, expected %q", line, "line 2")
	}
	if no != 2 {
		t.Errorf("line 2 (rescan): s.Line() = %d, expected %d", no, 2)
	}

	res = s.Scan()
	line, no = s.Text(), s.Line()
	if !res {
		t.Errorf("line 3: s.Scan() = false, expected true")
	}
	if line != "line 3" {
		t.Errorf("line 3: s.Text() = %q, expected %q", line, "line 3")
	}
	if no != 3 {
		t.Errorf("line 3: s.Line() = %d, expected %d", no, 3)
	}

	res = s.Scan()
	line, no = s.Text(), s.Line()
	if res {
		t.Errorf("line 4: s.Scan() = true, expected false")
	}
	if line != "" {
		t.Errorf("line 4: s.Text() = %q, expected %q", line, "")
	}
	if no != 4 {
		t.Errorf("line 4: s.Line() = %d, expected %d", no, 4)
	}
}
