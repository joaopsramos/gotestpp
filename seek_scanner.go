package main

import "bufio"

type SeekScanner struct {
	scanner        *bufio.Scanner
	prevLine       []byte
	returnPrevLine bool
}

func (s *SeekScanner) Scan() bool {
	s.returnPrevLine = false
	s.prevLine = s.scanner.Bytes()
	return s.scanner.Scan()
}

func (s *SeekScanner) Bytes() []byte {
	if s.returnPrevLine {
		return s.prevLine
	}

	return s.scanner.Bytes()
}

func (s *SeekScanner) Text() string {
	return string(s.Bytes())
}

func (s *SeekScanner) Seek() {
	s.returnPrevLine = true
}

func NewSeekScanner(scanner *bufio.Scanner) *SeekScanner {
	return &SeekScanner{scanner: scanner}
}
