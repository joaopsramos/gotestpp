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

		case <-errChan:
			return
		}
	}

	r.printFailedPkgs()
	r.printSkipped()
	r.printFailures()

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

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		if matches := errorFileRe.FindStringSubmatch(line); len(matches) > 0 {
			line = "\t" + color.CyanString(matches[1]) + color.RedString(matches[2])
			outputLines = append(outputLines, line)
			continue
		}

		if IsTestifyAssert(line) {
			testifyAssert := NewTestifyAssert(line, scanner)
			outputLines = append(outputLines, testifyAssert.String())
			continue
		}

		if t.Panicked && panicFileRe.MatchString(line) {
			line = "\t" + line
		}

		line = "\t\t" + color.RedString(line)
		outputLines = append(outputLines, line)
	}

	var output string
	failedSubTests := t.FilterSubTestsByAction("fail")

	if len(outputLines) > 0 {
		logs := r.formatLogs(t)
		if logs != "" {
			output += logs + "\n"
		}
		output += fmt.Sprintf(
			"%s %s (%.2fs)\n%s\n",
			color.RedString("--- FAIL"),
			t.Name,
			t.Elapsed,
			strings.Join(outputLines, "\n"),
		)

		if len(failedSubTests) > 0 {
			output += "\n"
		}
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
		output += fmt.Sprintf("%s %s", color.BlueString("Log:"), strings.TrimSpace(log))
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
