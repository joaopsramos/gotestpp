package main

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

type TestifyAssert struct {
	Error   []string
	Trace   []string
	Test    string
	Message string
}

func IsTestifyAssert(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "Error Trace:")
}

func NewTestifyAssert(firstLine string, scanner *RewindScanner) TestifyAssert {
	t := TestifyAssert{Trace: []string{firstLine}}
	isErrorTrace := true

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "Error:") {
			line = strings.TrimPrefix(line, "Error:")
			isErrorTrace = false
		}

		if isErrorTrace {
			t.Trace = append(t.Trace, line)
			continue
		}

		if strings.HasPrefix(line, "Test:") {
			t.Test = line
			continue
		}

		if strings.HasPrefix(line, "Messages:") {
			line = strings.TrimPrefix(line, "Messages:")
			t.Message = strings.TrimSpace(line)
			break
		}

		if t.Test != "" {
			// the testify output has ended, but we read an extra line,
			// so we need to rewind the scanner.
			scanner.Rewind()
			break
		}

		t.Error = append(t.Error, line)
	}

	return t
}

func (t TestifyAssert) String() string {
	return fmt.Sprintf("%s\n%s", t.formatError(), t.formatTrace())
}

func (t TestifyAssert) formatError() string {
	firstLine := color.RedString(t.Error[0])
	output := []string{firstLine}

	for _, line := range t.Error[1:] {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "-") {
			line = color.RedString(line)
		} else if strings.HasPrefix(line, "+") {
			line = color.GreenString(line)
		}

		output = append(output, line)
	}

	return fmt.Sprintf("\t%s\n\t%s", "Error:", strings.Join(output, "\n\t\t"))
}

func (t TestifyAssert) formatTrace() string {
	var output []string
	for _, line := range t.Trace {
		line = strings.TrimSpace(line)
		line = strings.Replace(line, "Error Trace:", "", 1)

		output = append(output, line)
	}

	return fmt.Sprintf("\t%s\n\t%s", "Error Trace:", strings.Join(output, "\n\t\t"))
}
