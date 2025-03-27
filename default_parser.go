package main

import (
	"bufio"
	"encoding/json"
	"io"
	"slices"
	"strings"
)

var actionsToIgnore = []string{"run", "start", "pause", "cont"}

type DefaultParser struct {
	testsMap    map[string]*TestEntry
	subTestsMap map[string][]*TestEntry
}

// Parse implements Parser.
func (p *DefaultParser) Parse(r io.Reader, testsChan chan<- TestEntry) error {
	scanner := NewSeekScanner(bufio.NewScanner(r))

	for scanner.Scan() {
		var logEntry LogEntry
		line := scanner.Bytes()
		json.Unmarshal(line, &logEntry)

		if slices.Contains(actionsToIgnore, logEntry.Action) {
			continue
		}

		test, ok := p.testsMap[logEntry.buildID()]
		if !ok {
			test = &TestEntry{Name: logEntry.Name, Pkg: logEntry.Pkg}
			p.testsMap[logEntry.buildID()] = test
		}

		test.Elapsed += logEntry.Elapsed

		if logEntry.Action != "output" {
			test.Action = logEntry.Action
		}

		switch logEntry.Action {
		case "pass", "skip":
			if test.IsPkg() {
				test.PkgFinished = true
			}

			testsChan <- *test

		case "fail":
			if test.IsPkg() {
				test.PkgFinished = true
			}
			test.PkgHasErrors = true

			if test.IsSubTest() {
				key := test.RootTestName()
				p.subTestsMap[key] = append(p.subTestsMap[key], test)
				continue
			}

			test.SubTests = p.getSubTests(test.Name)
			testsChan <- *test

		case "output":
			trimmed := strings.TrimSpace(logEntry.Output)

			if strings.HasPrefix(trimmed, "?") {
				test.NoTestFiles = true
				continue
			}

			if strings.HasPrefix(trimmed, "ok") {
				test.Cached = strings.Contains(trimmed, "(cached)")
				continue
			}

			if strings.HasPrefix(trimmed, "FAIL") {
				test.BuildFailed = strings.Contains(trimmed, "[build failed]")
				continue
			}

			if strings.HasPrefix(trimmed, "panic:") {
				test.Panicked = true
			}

			if test.Panicked ||
				strings.HasPrefix(logEntry.Output, " ") ||
				strings.HasPrefix(logEntry.Output, "\t") ||
				strings.HasPrefix(trimmed, "--- SKIP") {

				test.Output += logEntry.Output
				continue
			}

			firstPart := strings.Split(trimmed, " ")[0]
			if slices.Contains([]string{"---", "===", "PASS"}, firstPart) {
				continue
			}

			// Everything else is a log
			testsChan <- TestEntry{Action: "log", Output: logEntry.Output, Pkg: logEntry.Pkg}

		default:
			continue
		}
	}

	return nil
}

func (p *DefaultParser) getSubTests(key string) []TestEntry {
	tests := make([]TestEntry, len(p.subTestsMap[key]))
	for i, t := range p.subTestsMap[key] {
		tests[i] = *t
	}

	return tests
}

func NewDefaultParser() *DefaultParser {
	testsMap := make(map[string]*TestEntry)
	subTestsMap := make(map[string][]*TestEntry)

	return &DefaultParser{testsMap: testsMap, subTestsMap: subTestsMap}
}
