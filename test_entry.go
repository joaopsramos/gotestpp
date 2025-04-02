package main

import (
	"strings"
)

type TestEntry struct {
	Name         string
	Pkg          string
	Elapsed      float64
	Action       string
	Output       string
	SubTests     []TestEntry
	NoTestFiles  bool
	PkgFinished  bool
	PkgHasErrors bool
	Cached       bool
	BuildFailed  bool
	Panicked     bool
}

func (t TestEntry) RootTestName() string {
	rootTestName, _, _ := strings.Cut(t.Name, "/")
	return rootTestName
}

func (t TestEntry) IsSubTest() bool {
	return strings.Contains(t.Name, "/")
}

func (t TestEntry) IsPkg() bool {
	return t.Name == ""
}

func (t *TestEntry) FilterSubTestsByAction(action string) []TestEntry {
	result := []TestEntry{}
	for _, tt := range t.SubTests {
		if tt.Action == action {
			result = append(result, tt)
		}
	}

	return result
}
