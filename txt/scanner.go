package txt

import (
	"bufio"
	"io"
	"strings"
)

// scanner behaves very similar to [bufio.Scanner] but offers additional
// functionality of undoing a single scan, thereby repeating the line last read.
type scanner struct {
	scanner  *bufio.Scanner
	prevLine string
	reset    bool
	lineNo   int
}

func newScanner(r io.Reader) *scanner {
	return &scanner{
		scanner:  bufio.NewScanner(r),
		prevLine: "",
		reset:    false,
		lineNo:   0,
	}
}

func (s *scanner) scan() bool {
	s.lineNo++
	if s.reset {
		s.reset = false
		return true
	}
	s.prevLine = s.scanner.Text()
	res := s.scanner.Scan()
	return res
}

func (s *scanner) unScan() {
	if s.lineNo <= 0 {
		panic("unScan called before scan")
	}
	s.lineNo--
	s.reset = true
}

func (s *scanner) skipEmptyLines() error {
	for s.scan() {
		if strings.TrimSpace(s.text()) != "" {
			s.unScan()
			break
		}
	}
	return s.err()
}

func (s *scanner) text() string {
	if s.reset {
		return s.prevLine
	}
	return s.scanner.Text()
}

func (s *scanner) err() error {
	return s.scanner.Err()
}

func (s *scanner) line() int {
	return s.lineNo
}
