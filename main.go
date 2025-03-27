package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
)

type Parser interface {
	Parse(reader io.Reader, testsChan chan<- TestEntry) error
}

type OutputRenderer interface {
	Render(testsChan <-chan TestEntry)
}

func main() {
	args := os.Args[1:]
	args = append([]string{"test", "-json"}, args...)
	parser := NewDefaultParser()
	renderer := NewDefaultOutputRenderer()

	os.Exit(run(args, parser, renderer))
}

func run(args []string, parser Parser, renderer OutputRenderer) int {
	var wg sync.WaitGroup
	wg.Add(1)

	r, w := io.Pipe()

	cmd := exec.Command("go", args...)
	cmd.Stderr = w
	cmd.Stdout = w
	cmd.Env = os.Environ()

	if err := cmd.Start(); err != nil {
		log.Println(err)
		wg.Done()
		return 1
	}

	go process(&wg, r, parser, renderer)

	cmd.Wait()
	w.Close()
	wg.Wait()

	return cmd.ProcessState.ExitCode()
}

func process(wg *sync.WaitGroup, r io.Reader, parser Parser, renderer OutputRenderer) {
	defer wg.Done()

	testsChan := make(chan TestEntry)

	go func() {
		defer close(testsChan)

		err := parser.Parse(r, testsChan)
		if err != nil {
			log.Println(err)
			return
		}
	}()

	renderer.Render(testsChan)
}
