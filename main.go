package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
)

func main() {
	args := os.Args[1:]
	args = append([]string{"test", "-json"}, args...)
	parser := NewParser()
	renderer := NewRenderer()

	os.Exit(run(args, parser, renderer))
}

func run(args []string, parser *Parser, renderer *Renderer) int {
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

func process(wg *sync.WaitGroup, r io.Reader, parser *Parser, renderer *Renderer) {
	defer wg.Done()

	testsChan := make(chan TestEntry)
	errChan := make(chan error)

	go func() {
		defer close(testsChan)
		defer close(errChan)

		parser.Parse(r, testsChan, errChan)
	}()

	renderer.Render(testsChan, errChan)
}
