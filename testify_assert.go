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
	Message []string
}

func IsTestifyAssert(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "Error Trace:")
}

func NewTestifyAssert(firstLine string, scanner *RewindScanner) TestifyAssert {
	t := TestifyAssert{Trace: []string{strings.TrimSpace(firstLine)}}
	isErrorTrace := true
	errorStarted := false
	messageStarted := false

Loop:
	for scanner.Scan() {
		nonTrimmed := scanner.Text()
		line := strings.TrimSpace(nonTrimmed)

		if strings.HasPrefix(line, "Error:") && !errorStarted {
			isErrorTrace = false
			errorStarted = true
			line = strings.TrimPrefix(line, "Error:")
			t.Error = append(t.Error, strings.TrimSpace(line))
			continue
		}

		if strings.HasPrefix(line, "Messages:") && !messageStarted {
			messageStarted = true
			line = strings.TrimPrefix(line, "Messages:")
			t.Message = append(t.Message, strings.TrimSpace(line))
			continue
		}

		switch {
		case isErrorTrace:
			t.Trace = append(t.Trace, line)

		case strings.HasPrefix(line, "Test:"):
			t.Test = line

		case messageStarted &&
			errorFileRe.Match(scanner.Bytes()) &&
			// Current line identation is lte than first line
			t.countSpaces(nonTrimmed) <= t.countSpaces(firstLine):

			// Message has ended, but we already read the next line, so we need to rewind the scanner.
			scanner.Rewind()
			break Loop

		case messageStarted:
			t.Message = append(t.Message, line)

		case t.Test != "" && !messageStarted:
			// The testify output has ended with no message, but we already read the next line,
			// so we need to rewind the scanner.
			scanner.Rewind()
			break Loop

		default:
			// TODO: Keep identation of diff lines
			t.Error = append(t.Error, line)
		}
	}

	return t
}

func (t TestifyAssert) String() string {
	output := t.formatError()
	if len(t.Message) > 0 {
		output += fmt.Sprintf("\n\tMessages:\n\t\t%s", strings.Join(t.Message, "\n\t\t"))
	}
	output += "\n" + t.formatTrace()

	return output
}

func (t TestifyAssert) formatError() string {
	firstLine := color.RedString(t.Error[0])
	output := []string{firstLine}

	for _, line := range t.Error[1:] {
		if strings.HasPrefix(line, "-") {
			line = color.RedString(line)
		} else if strings.HasPrefix(line, "+") {
			line = color.GreenString(line)
		}

		output = append(output, line)
	}

	return fmt.Sprintf("\t%s\n\t\t%s", "Error:", strings.Join(output, "\n\t\t"))
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

func (t TestifyAssert) countSpaces(s string) int {
	count := 0
Loop:
	for _, c := range s {
		switch c {
		case ' ':
			count++
		case '\t':
			count += 4
		default:
			break Loop
		}
	}

	return count
}
