package main

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/fatih/color"
)

var (
	blue   = color.New(color.FgBlue)
	yellow = color.New(color.FgYellow)
	red    = color.New(color.FgRed)
)

type DefaultOutputRenderer struct {
	summary        summary
	logsByPkg      map[string]pkgLogs
	failedPkgs     []string
	failedOutputs  []string
	skippedOutputs []string
}

type summary struct {
	Passed  int
	Failed  int
	Skipped int
	Elapsed float64
}

type pkgLogs struct {
	pkgHasErrors bool
	logs         []string
}

func NewDefaultOutputRenderer() *DefaultOutputRenderer {
	return &DefaultOutputRenderer{logsByPkg: make(map[string]pkgLogs)}
}

// Render implements OutputRenderer.
func (r *DefaultOutputRenderer) Render(testsChan <-chan TestEntry) {
	for t := range testsChan {
		switch t.Action {
		case "log":
			r.handleLog(t)

		case "pass":
			r.handlePass(t)

		case "skip":
			r.handleSkip(t)

		case "fail":
			r.handleFail(t)
		}
	}

	r.printFailedPkgs()
	r.printLogs()
	r.printSkipped()
	r.printFailures()

	fmt.Println()
	r.printSummary()
	fmt.Println()
}

func (r *DefaultOutputRenderer) handleLog(t TestEntry) {
	l := r.logsByPkg[t.Pkg]
	l.logs = append(l.logs, t.Output)
	r.logsByPkg[t.Pkg] = l
}

func (r *DefaultOutputRenderer) handlePass(t TestEntry) {
	if !t.IsPkg() {
		r.summary.Passed++
		r.summary.Elapsed += t.Elapsed
	}

	if t.Cached {
		fmt.Printf("%s\t%s\t(cached)\n", color.GreenString("ok"), t.Pkg)
		return
	}

	if t.PkgFinished {
		fmt.Printf("%s\t%s\t%.2fs\n", color.GreenString("ok"), t.Pkg, t.Elapsed)
	}
}

func (r *DefaultOutputRenderer) handleSkip(t TestEntry) {
	if t.NoTestFiles {
		fmt.Printf("%s\t%s\t[no test files]\n", color.YellowString("?"), t.Pkg)
		return
	}

	r.summary.Skipped++

	file := strings.TrimSpace(strings.Split(t.Output, "\n")[0])

	output := yellow.Sprintf("%s %s (%.2fs)\n", "--- SKIP", t.Name, t.Elapsed)
	if file != "" {
		output += yellow.Sprintf("\t%s\n", strings.TrimSuffix(file, ":"))
	}

	r.skippedOutputs = append(r.skippedOutputs, output)
}

func (r *DefaultOutputRenderer) handleFail(t TestEntry) {
	if t.BuildFailed {
		r.failedPkgs = append(r.failedPkgs, red.Sprintf("FAIL\t%s\t[build failed]\n", t.Pkg))
		return
	}

	if t.IsPkg() {
		if l, ok := r.logsByPkg[t.Pkg]; ok {
			l.pkgHasErrors = true
			r.logsByPkg[t.Pkg] = l
		}

		output := fmt.Sprintf("%s\t%s\n", color.RedString("FAIL"), t.Pkg)
		r.failedPkgs = append(r.failedPkgs, output)
		return
	}

	r.summary.Elapsed += t.Elapsed
	r.summary.Failed += 1 + len(t.SubTests)
	r.failedOutputs = append(r.failedOutputs, r.formatError(t))
}

func (r *DefaultOutputRenderer) formatError(entry TestEntry) string {
	outputLines := []string{}
	reader := strings.NewReader(entry.Output)
	scanner := NewSeekScanner(bufio.NewScanner(reader))

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		matches := errorFileRegex.FindStringSubmatch(trimmed)

		if len(matches) > 0 {
			line = "\t" + color.CyanString(matches[1]) + color.RedString(matches[2])
			outputLines = append(outputLines, line)
			continue
		}

		if strings.HasPrefix(trimmed, "Error Trace:") {
			errorOutput := r.formatTestifyAssert(line, scanner)
			outputLines = append(outputLines, errorOutput)
			continue
		}

		if line != "" {
			outputLines = append(outputLines, color.RedString(line))
		}
	}

	var output string

	if len(outputLines) > 0 {
		output = fmt.Sprintf(
			"%s %s (%.2fs)\n%s\n",
			color.RedString("--- FAIL"),
			entry.Name,
			entry.Elapsed,
			strings.Join(outputLines, "\n"),
		)

		if len(entry.SubTests) > 0 {
			output += "\n"
		}
	}

	subTestsOutput := make([]string, len(entry.SubTests))
	for i, st := range entry.SubTests {
		subTestsOutput[i] = r.formatError(st)
	}

	output += strings.Join(subTestsOutput, "\n")

	return output
}

func (r *DefaultOutputRenderer) formatTestifyAssert(line string, scanner *SeekScanner) string {
	isErrorTrace := true
	errorTraceLines := []string{line}

	var errorLines []string

	for scanner.Scan() {
		line = scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "Test:") {
			scanner.Seek()
			break
		}

		if strings.HasPrefix(trimmed, "Error:") {
			isErrorTrace = false
		}

		if isErrorTrace {
			errorTraceLines = append(errorTraceLines, line)
			continue
		}

		errorLines = append(errorLines, line)
	}

	return fmt.Sprintf(
		"%s\n%s",
		r.formatTestifyError(errorLines),
		r.formatTestifyTrace(errorTraceLines),
	)
}

func (r *DefaultOutputRenderer) formatTestifyError(lines []string) string {
	var formattedLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Error:") {
			line = "\t" + strings.Replace(line, "Error:", "", 1)
			line = color.RedString(line)
		}

		if strings.HasPrefix(line, "-") {
			line = color.RedString(line)
		} else if strings.HasPrefix(line, "+") {
			line = color.GreenString(line)
		}

		formattedLines = append(formattedLines, line)
	}

	formattedOutput := strings.Join(formattedLines, "\n\t\t")
	return fmt.Sprintf("\t%s\n%s", "Error:", formattedOutput)
}

func (r *DefaultOutputRenderer) formatTestifyTrace(trace []string) string {
	var formattedLines []string
	for _, line := range trace {
		line = strings.TrimSpace(line)
		line = strings.Replace(line, "Error Trace:", "", 1)

		formattedLines = append(formattedLines, line)
	}

	formattedOutput := strings.Join(formattedLines, "\n\t\t")
	return fmt.Sprintf("\t%s\n\t%s", "Error Trace:", formattedOutput)
}

func (r DefaultOutputRenderer) printFailedPkgs() {
	fmt.Print(strings.Join(r.failedPkgs, ""))
}

func (r DefaultOutputRenderer) printLogs() {
	for _, pkgLog := range r.logsByPkg {
		if !pkgLog.pkgHasErrors {
			continue
		}

		for _, log := range pkgLog.logs {
			fmt.Printf("%s %s", color.BlueString("Log:"), log)
		}
	}
}

func (r DefaultOutputRenderer) printSkipped() {
	if len(r.skippedOutputs) > 0 {
		fmt.Print("\n" + strings.Join(r.skippedOutputs, "\n"))
	}
}

func (r DefaultOutputRenderer) printFailures() {
	if len(r.failedOutputs) > 0 {
		fmt.Print("\n" + strings.Join(r.failedOutputs, "\n"))
	}
}

func (r DefaultOutputRenderer) printSummary() {
	s := r.summary

	output := fmt.Sprintf("%d tests", s.Passed+s.Failed+s.Skipped)

	if s.Failed > 0 {
		output = red.Sprintf("%s, %d failed", output, s.Failed)
	} else {
		output = color.GreenString(output)
	}

	if s.Skipped > 0 {
		if s.Failed > 0 {
			output += color.RedString(", ")
		} else {
			output += color.GreenString(", ")
		}

		output += yellow.Sprintf("%d skipped", s.Skipped)
	}

	fmt.Printf("Finished in %.2fs\n%s", s.Elapsed, output)
}
