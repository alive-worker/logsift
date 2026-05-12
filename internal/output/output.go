package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/alive-worker/logsift/internal/parser"
)

type Writer interface {
	Write(e *parser.Entry) error
}

func New(format string, w io.Writer, useColor bool) (Writer, error) {
	switch format {
	case "", "color":
		return &colorWriter{w: w, color: useColor}, nil
	case "json":
		return &jsonWriter{w: w}, nil
	case "tsv":
		return &tsvWriter{w: w}, nil
	}
	return nil, fmt.Errorf("unknown output format %q", format)
}

type colorWriter struct {
	w     io.Writer
	color bool
}

func (c *colorWriter) Write(e *parser.Entry) error {
	level := strings.ToUpper(e.Level)
	if c.color {
		level = colorize(e.Level, level)
	}
	ts := "????"
	if !e.Timestamp.IsZero() {
		ts = e.Timestamp.Format("15:04:05")
	}
	_, err := fmt.Fprintf(c.w, "%s %-5s %-12s %s\n", ts, level, e.Service, e.Message)
	return err
}

func colorize(level, text string) string {
	codes := map[string]string{
		"error": "31", "err": "31",
		"warn": "33", "warning": "33",
		"info":  "36",
		"debug": "37",
	}
	c, ok := codes[strings.ToLower(level)]
	if !ok {
		return text
	}
	return "\x1b[" + c + "m" + text + "\x1b[0m"
}

type jsonWriter struct{ w io.Writer }

func (j *jsonWriter) Write(e *parser.Entry) error {
	_, err := fmt.Fprintln(j.w, e.Raw)
	return err
}

type tsvWriter struct{ w io.Writer }

func (t *tsvWriter) Write(e *parser.Entry) error {
	ts := ""
	if !e.Timestamp.IsZero() {
		ts = e.Timestamp.Format("2006-01-02T15:04:05Z07:00")
	}
	_, err := fmt.Fprintf(t.w, "%s\t%s\t%s\t%s\n", ts, e.Level, e.Service, e.Message)
	return err
}
