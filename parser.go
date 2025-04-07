package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"slices"
	"strings"
)

const test2jsonOutBuffer = 1024

var actionsToIgnore = []string{"run", "start", "pause", "cont"}

type Parser struct {
	testsMap    map[string]*TestEntry
	subTestsMap map[string][]*TestEntry
}

func (p *Parser) Parse(r io.Reader, testsChan chan<- TestEntry, errsChan chan<- error) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		var event TestEvent
		line := scanner.Bytes()
		err := json.Unmarshal(line, &event)
		if err != nil {
			errsChan <- errors.New(string(line))
			continue
		}

		if slices.Contains(actionsToIgnore, event.Action) {
			continue
		}

		eventID := event.buildID()
		test, ok := p.testsMap[eventID]
		if !ok {
			test = &TestEntry{Name: event.Name, Pkg: event.Pkg, EventID: eventID}
			p.testsMap[eventID] = test
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

			if test.IsSubTest() {
				key := test.RootTestName()
				p.subTestsMap[key] = append(p.subTestsMap[key], test)
				continue
			}

			p.sendTest(test, testsChan)

		case "output":
			if strings.HasPrefix(event.Output, "panic:") {
				test.Panicked = true
			}

			switch {
			case p.ignoreOutput(event.Output):
				continue

			case strings.HasPrefix(event.Output, "?"):
				test.NoTestFiles = true

			case strings.HasPrefix(event.Output, "ok"):
				test.Cached = strings.Contains(event.Output, "(cached)")

			case strings.HasPrefix(event.Output, "FAIL"):
				test.BuildFailed = strings.Contains(event.Output, "[build failed]")

			default:
				test.Output += event.Output
			}

		default:
			continue
		}
	}

	// Send remaining events that don't have a pass/skip/fail action
	for _, test := range p.testsMap {
		if strings.TrimSpace(test.Output) != "" {
			testsChan <- *test
		}
	}
}

func (p *Parser) ignoreOutput(output string) bool {
	for _, prefix := range []string{"===", "--- PASS", "--- SKIP", "--- FAIL", "PASS"} {
		if strings.HasPrefix(output, prefix) {
			return true
		}
	}

	return false
}

func (p *Parser) sendTest(test *TestEntry, testsChan chan<- TestEntry) {
	test.SubTests = p.getSubTests(test.RootTestName())
	testsChan <- *test
	p.deleteTest(test)
}

func (p *Parser) getSubTests(key string) []TestEntry {
	tests := make([]TestEntry, len(p.subTestsMap[key]))
	for i, t := range p.subTestsMap[key] {
		tests[i] = *t
	}

	return tests
}

func (p *Parser) deleteTest(test *TestEntry) {
	delete(p.testsMap, test.EventID)

	for _, subTest := range test.SubTests {
		delete(p.testsMap, subTest.EventID)
	}
}

func NewParser() *Parser {
	testsMap := make(map[string]*TestEntry)
	subTestsMap := make(map[string][]*TestEntry)

	return &Parser{testsMap: testsMap, subTestsMap: subTestsMap}
}
