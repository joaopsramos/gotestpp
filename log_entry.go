package main

import (
	"strings"
	"time"
)

type LogEntry struct {
	Time    time.Time
	Action  string
	Pkg     string `json:"Package"`
	Name    string `json:"Test"`
	Output  string
	Elapsed float64
}

func (l LogEntry) buildID() string {
	return strings.Join([]string{l.Pkg, l.Name}, "-")
}
