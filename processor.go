package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
)

type Processor struct {
	parser   *Parser
	renderer *Renderer
}

func NewProcessor() *Processor {
	return &Processor{parser: NewParser(), renderer: NewRenderer()}
}

func (p *Processor) Run() int {
	if stat, _ := os.Stdin.Stat(); (stat.Mode() & os.ModeNamedPipe) != 0 {
		return p.Process(os.Stdin)
	}

	return p.runWithCmd()
}

func (p *Processor) Process(r io.Reader) int {
	testsChan := make(chan TestEntry)
	errChan := make(chan error)

	go func() {
		defer close(testsChan)
		defer close(errChan)

		p.parser.Parse(r, testsChan, errChan)
	}()

	err := p.renderer.Render(testsChan, errChan)
	if err != nil {
		return 1
	}

	return 0
}

func (p *Processor) runWithCmd() int {
	var wg sync.WaitGroup
	wg.Add(1)

	r, w := io.Pipe()

	args := os.Args[1:]
	args = append([]string{"test", "-json"}, args...)

	cmd := exec.Command("go", args...)
	cmd.Stderr = w
	cmd.Stdout = w
	cmd.Env = os.Environ()

	if err := cmd.Start(); err != nil {
		log.Println(err)
		wg.Done()
		return 1
	}

	result := make(chan int, 1)

	go func() {
		defer wg.Done()
		result <- p.Process(r)
	}()

	cmd.Wait()
	w.Close()
	wg.Wait()

	if cmd.ProcessState.ExitCode() == 0 {
		return <-result
	}

	return cmd.ProcessState.ExitCode()
}
