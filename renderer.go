package main

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

var (
	blue   = color.New(color.FgBlue)
	yellow = color.New(color.FgYellow)
	red    = color.New(color.FgRed)

	errorFileRe = regexp.MustCompile(`^([\w\s.-]+\.go:\d+:)(.*)`)
	panicFileRe = regexp.MustCompile(`^([a-zA-Z]:\\|/)?([\w\s.-]+[/\\])*[\w\s.-]+\.go:\d+`)
)

type Renderer struct {
	summary        Summary
	failedPkgs     []string
	failedOutputs  []string
	skippedOutputs []string
	errors         []string
}

type pkgLogs struct {
	pkgHasErrors bool
	logs         []string
}

func NewRenderer() *Renderer {
	return &Renderer{}
}

func (r *Renderer) Render(testsChan <-chan TestEntry, errChan <-chan error) {
Loop:
	for {
		select {
		case t, ok := <-testsChan:
			if !ok {
				break Loop
			}

			switch t.Action {
			case "pass":
				r.handlePass(t)

			case "skip":
				r.handleSkip(t)

			case "fail":
				r.handleFail(t)
			}

		case err, ok := <-errChan:
			if !ok {
				break Loop
			}

			r.errors = append(r.errors, err.Error())
		}
	}

	r.printFailedPkgs()
	r.printSkipped()
	r.printFailures()
	r.printErrors()

	fmt.Printf("\n%s\n", r.summary)
}

func (r *Renderer) handlePass(t TestEntry) {
	if !t.IsPkg() {
		r.summary.Passed += 1 + len(t.FilterSubTestsByAction("pass"))
		return
	}

	if t.Cached {
		fmt.Printf("%s\t%s\t(cached)\n", color.GreenString("ok"), t.Pkg)
		return
	}

	r.summary.Elapsed += t.Elapsed

	if t.PkgFinished {
		fmt.Printf("%s\t%s\t%.2fs\n", color.GreenString("ok"), t.Pkg, t.Elapsed)
	}
}

func (r *Renderer) handleSkip(t TestEntry) {
	if t.NoTestFiles {
		fmt.Printf("%s\t%s\t[no test files]\n", color.YellowString("?"), t.Pkg)
		return
	}

	r.summary.Elapsed += t.Elapsed
	r.summary.Skipped++

	file := strings.TrimSpace(strings.Split(t.Output, "\n")[0])

	output := yellow.Sprintf("%s %s (%.2fs)\n", "--- SKIP", t.Name, t.Elapsed)
	if file != "" {
		output += yellow.Sprintf("\t%s\n", strings.TrimSuffix(file, ":"))
	}

	r.skippedOutputs = append(r.skippedOutputs, output)
}

func (r *Renderer) handleFail(t TestEntry) {
	if t.BuildFailed {
		r.failedPkgs = append(r.failedPkgs, red.Sprintf("FAIL\t%s\t[build failed]\n", t.Pkg))
		return
	}

	if t.IsPkg() {
		r.summary.Elapsed += t.Elapsed

		output := fmt.Sprintf("%s\t%s\n", color.RedString("FAIL"), t.Pkg)
		r.failedPkgs = append(r.failedPkgs, output)
		return
	}

	r.summary.Failed += 1 + len(t.FilterSubTestsByAction("fail"))
	r.summary.Passed += len(t.FilterSubTestsByAction("pass"))
	r.failedOutputs = append(r.failedOutputs, r.formatError(t))
}

func (r *Renderer) formatError(t TestEntry) string {
	outputLines := []string{}
	reader := strings.NewReader(t.Output)
	scanner := NewRewindScanner(bufio.NewScanner(reader))
	panicStarted := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "panic:") {
			panicStarted = true
		}

		switch {
		case line == "":
			continue

		case IsTestifyAssert(line):
			nonTrimmed := scanner.Text()
			testifyAssert := NewTestifyAssert(nonTrimmed, scanner)
			outputLines = append(outputLines, testifyAssert.String())

		case t.Panicked && panicStarted:
			if panicFileRe.MatchString(line) {
				line = "\t" + line
			}
			outputLines = append(outputLines, color.RedString(line))

		default:
			if matches := errorFileRe.FindStringSubmatch(line); len(matches) > 0 {
				line = "\t" + color.CyanString(matches[1]) + color.RedString(matches[2])
				outputLines = append(outputLines, line)
				continue
			}

			line = "\t\t" + color.RedString(line)
			outputLines = append(outputLines, line)
		}
	}

	var output string
	failedSubTests := t.FilterSubTestsByAction("fail")

	output += r.formatLogs(t)
	output += fmt.Sprintf("%s %s (%.2fs)\n", color.RedString("--- FAIL"), t.Name, t.Elapsed)

	if len(outputLines) > 0 {
		output += fmt.Sprintf("%s\n", strings.Join(outputLines, "\n"))
	}

	if len(failedSubTests) > 0 && len(outputLines) > 0 {
		output += "\n"
	}

	subTestsOutput := make([]string, len(failedSubTests))
	for i, st := range failedSubTests {
		subTestsOutput[i] = r.formatError(st)
	}

	output += strings.Join(subTestsOutput, "\n")

	return output
}

func (r Renderer) formatLogs(t TestEntry) string {
	var output string
	for _, log := range t.Logs {
		output += fmt.Sprintf("%s %s\n", color.BlueString("Log:"), strings.TrimSpace(log))
	}

	return output
}

func (r Renderer) printFailedPkgs() {
	fmt.Print(strings.Join(r.failedPkgs, ""))
}

func (r Renderer) printSkipped() {
	if len(r.skippedOutputs) > 0 {
		fmt.Print("\n" + strings.Join(r.skippedOutputs, "\n"))
	}
}

func (r Renderer) printFailures() {
	if len(r.failedOutputs) > 0 {
		fmt.Print("\n" + strings.Join(r.failedOutputs, "\n"))
	}
}

func (r Renderer) printErrors() {
	if len(r.errors) > 0 {
		fmt.Printf("\n%s\n%s\n", color.RedString("Errors:"), strings.Join(r.errors, "\n"))
	}
}
