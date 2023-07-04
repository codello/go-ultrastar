package txt

import (
	"bufio"
	"io"
	"strings"
)

// scanner behaves very similar to [bufio.Scanner] but offers additional
// functionality of undoing a single scan, thereby repeating the line last read.
type scanner struct {
	scanner     *bufio.Scanner
	lineNo      int
	usePrevLine bool
	prevLine    string
	prevLineNo  int

	SkipEmptyLines        bool
	TrimLeadingWhitespace bool
}

func newScanner(r io.Reader) *scanner {
	return &scanner{
		scanner:     bufio.NewScanner(r),
		lineNo:      0,
		usePrevLine: false,
		prevLine:    "",
		prevLineNo:  -1,
	}
}

func (s *scanner) Scan() bool {
	if s.usePrevLine {
		s.usePrevLine = false
		return true
	}
	s.prevLine = s.scanner.Text()
	s.prevLineNo = s.lineNo
	res := s.scanner.Scan()
	s.lineNo++

	if s.SkipEmptyLines {
		for res && strings.TrimSpace(s.scanner.Text()) == "" {
			res = s.scanner.Scan()
			s.lineNo++
		}
	}
	return res
}

func (s *scanner) UnScan() {
	if s.prevLineNo < 0 {
		panic("UnScan called before scan.")
	}
	s.usePrevLine = true
}

func (s *scanner) ScanEmptyLines() error {
	// TODO: Doc: Invalidates unScan. If SkipEmptyLines is true this does basically nothing
	if s.usePrevLine && strings.TrimSpace(s.prevLine) != "" {
		return nil
	}
	for s.Scan() {
		if strings.TrimSpace(s.Text()) != "" {
			s.UnScan()
			return nil
		}
	}
	return s.Err()
}

func (s *scanner) Text() (text string) {
	if s.usePrevLine {
		text = s.prevLine
	} else {
		text = s.scanner.Text()
	}
	if s.TrimLeadingWhitespace {
		text = strings.TrimLeft(text, " \t")
	}
	return text
}

func (s *scanner) Err() error {
	return s.scanner.Err()
}

func (s *scanner) Line() int {
	if s.usePrevLine {
		return s.prevLineNo
	} else {
		return s.lineNo
	}
}
