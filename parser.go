package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"slices"
	"strings"
)

var actionsToIgnore = []string{"run", "start", "pause", "cont"}

type Parser struct {
	testsMap    map[string]*TestEntry
	subTestsMap map[string][]*TestEntry
}

func (p *Parser) Parse(r io.Reader, testsChan chan<- TestEntry) error {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		var event TestEvent
		line := scanner.Bytes()
		err := json.Unmarshal(line, &event)
		if err != nil {
			return errors.New(string(line))
		}

		if slices.Contains(actionsToIgnore, event.Action) {
			continue
		}

		test, ok := p.testsMap[event.buildID()]
		if !ok {
			test = &TestEntry{Name: event.Name, Pkg: event.Pkg}
			p.testsMap[event.buildID()] = test
		}

		test.Elapsed += event.Elapsed

		if event.Action != "output" {
			test.Action = event.Action
		}

		switch event.Action {
		case "pass", "skip", "fail":
			if test.IsPkg() {
				test.PkgFinished = true
			}

			test.PkgHasErrors = event.Action == "fail"
			key := test.RootTestName()

			if test.IsSubTest() {
				p.subTestsMap[key] = append(p.subTestsMap[key], test)
				continue
			}

			test.SubTests = p.getSubTests(key)
			testsChan <- *test

		case "output":
			trimmed := strings.TrimSpace(event.Output)

			if p.ignoreOutput(trimmed) {
				continue
			}

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
				strings.HasPrefix(event.Output, " ") ||
				strings.HasPrefix(event.Output, "\t") ||
				strings.HasPrefix(trimmed, "--- SKIP") {

				test.Output += event.Output
				continue
			}

			// Everything else is a log
			test.Logs = append(test.Logs, event.Output)

		default:
			continue
		}
	}

	return nil
}

func (p *Parser) ignoreOutput(trimmedOutput string) bool {
	for _, prefix := range []string{"===", "--- FAIL", "--- PASS", "PASS"} {
		if strings.HasPrefix(trimmedOutput, prefix) {
			return true
		}
	}

	return false
}

func (p *Parser) getSubTests(key string) []TestEntry {
	tests := make([]TestEntry, len(p.subTestsMap[key]))
	for i, t := range p.subTestsMap[key] {
		tests[i] = *t
	}

	return tests
}

func NewParser() *Parser {
	testsMap := make(map[string]*TestEntry)
	subTestsMap := make(map[string][]*TestEntry)

	return &Parser{testsMap: testsMap, subTestsMap: subTestsMap}
}
