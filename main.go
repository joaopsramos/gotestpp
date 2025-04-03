package main

import (
	"os"
)

func main() {
	processor := NewProcessor()
	result := processor.Run()

	os.Exit(result)
}
