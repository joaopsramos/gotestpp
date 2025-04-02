package main

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/joaopsramos/gotestpp/utils"
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
	isError := false
	isMessage := false

	firstLineIndent := utils.CountSpacesAndTabs(firstLine)

Loop:
	for scanner.Scan() {
		originalLine := scanner.Text()
		line := strings.TrimSpace(originalLine)
		indent := utils.CountSpacesAndTabs(originalLine)
		sameIndent := indent == firstLineIndent

		switch {
		case sameIndent && strings.HasPrefix(line, "Error:"):
			isErrorTrace = false
			isError = true
			t.Error = append(t.Error, originalLine)

		case sameIndent && strings.HasPrefix(line, "Messages:"):
			isError = false
			isMessage = true
			t.Message = append(t.Message, originalLine)

		case isErrorTrace:
			t.Trace = append(t.Trace, line)

		case strings.HasPrefix(line, "Test:"):
			isError = false
			t.Test = line

		case isError:
			t.Error = append(t.Error, originalLine)

		case isMessage && indent <= utils.CountSpacesAndTabs(firstLine):
			// Message has ended, but we already read the next line, so we need to rewind the scanner.
			scanner.Rewind()
			break Loop

		case isMessage:
			t.Message = append(t.Message, originalLine)

		case t.Test != "" && !isMessage:
			// The testify output has ended with no message, but we already read the next line,
			// so we need to rewind the scanner.
			scanner.Rewind()
			break Loop
		}
	}

	return t
}

func (t TestifyAssert) String() string {
	output := t.formatError()
	output += t.formatMessages()
	output += "\n" + t.formatTrace()

	return output
}

func (t TestifyAssert) formatError() string {
	output := make([]string, 0, len(t.Error))

	// Replace to get the correct indentation count
	firstLine := strings.Replace(t.Error[0], "Error:", "      ", 1)
	baseIndent := utils.CountSpacesAndTabs(firstLine)

	output = append(output, color.RedString(strings.TrimSpace(firstLine)))

	for _, line := range t.Error[1:] {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			output = append(output, "")
			continue
		}

		indent := utils.CountSpacesAndTabs(line)
		sameIndent := indent == baseIndent

		line = utils.StripExtraSpacesAndTabs(line)

		if indent == baseIndent+1 {
			// We removed a space that was part of the Diff context (without -/+ signs)
			line = " " + line
		}

		if sameIndent && strings.HasPrefix(trimmed, "-") {
			line = color.RedString(line)
		} else if sameIndent && strings.HasPrefix(trimmed, "+") {
			line = color.GreenString(line)
		}

		output = append(output, line)
	}

	return fmt.Sprintf("\t%s\n\t\t%s", "Error:", strings.Join(output, "\n\t\t"))
}

func (t TestifyAssert) formatMessages() string {
	if len(t.Message) == 0 {
		return ""
	}

	output := make([]string, 0, len(t.Message))

	t.Message[0] = strings.Replace(t.Message[0], "Messages:", "         ", 1)
	for _, line := range t.Message {
		line = utils.StripExtraSpacesAndTabs(line)
		output = append(output, line)
	}

	return fmt.Sprintf("\n\tMessages:\n\t\t%s", strings.Join(output, "\n\t\t"))
}

func (t TestifyAssert) formatTrace() string {
	t.Trace[0] = strings.Replace(t.Trace[0], "Error Trace:", "", 1)

	return fmt.Sprintf("\t%s\n\t%s", "Error Trace:", strings.Join(t.Trace, "\n\t\t"))
}
