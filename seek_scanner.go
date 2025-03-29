package main

import "bufio"

type RewindScanner struct {
	Scanner        *bufio.Scanner
	prevLine       []byte
	returnPrevLine bool
}

func (s *RewindScanner) Scan() bool {
	if s.returnPrevLine {
		s.returnPrevLine = false
		return true
	}

	return s.Scanner.Scan()
}

func (s *RewindScanner) Bytes() []byte {
	return s.Scanner.Bytes()
}

func (s *RewindScanner) Text() string {
	return string(s.Bytes())
}

func (s *RewindScanner) Rewind() {
	s.returnPrevLine = true
}

func NewRewindScanner(scanner *bufio.Scanner) *RewindScanner {
	return &RewindScanner{Scanner: scanner}
}
