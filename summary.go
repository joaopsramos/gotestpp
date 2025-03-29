package main

import (
	"fmt"

	"github.com/fatih/color"
)

type Summary struct {
	Passed  int
	Failed  int
	Skipped int
	Elapsed float64
}

func (s Summary) String() string {
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

	return fmt.Sprintf("Finished in %.2fs\n%s", s.Elapsed, output)
}
