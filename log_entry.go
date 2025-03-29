package main

import (
	"strings"
	"time"
)

type TestEvent struct {
	Time    time.Time
	Action  string
	Pkg     string `json:"Package"`
	Name    string `json:"Test"`
	Output  string
	Elapsed float64
}

func (l TestEvent) buildID() string {
	return strings.Join([]string{l.Pkg, l.Name}, "-")
}
