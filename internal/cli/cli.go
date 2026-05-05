package cli

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/alive-worker/logsift/internal/filter"
	"github.com/alive-worker/logsift/internal/output"
	"github.com/alive-worker/logsift/internal/parser"
)

const Version = "0.2.0"

type Options struct {
	File        string
	Level       string
	Since       string
	Grep        string
	Where       multiFlag
	Output      string
	ShowVersion bool
}

type multiFlag []string

func (m *multiFlag) String() string     { return fmt.Sprint(*m) }
func (m *multiFlag) Set(v string) error { *m = append(*m, v); return nil }

func ParseArgs(args []string, stderr io.Writer) (*Options, error) {
	fs := flag.NewFlagSet("logsift", flag.ContinueOnError)
	fs.SetOutput(stderr)
	opts := &Options{}
	fs.StringVar(&opts.File, "file", "", "path to NDJSON log file (default stdin)")
	fs.StringVar(&opts.Level, "level", "", "csv of levels to keep")
	fs.StringVar(&opts.Since, "since", "", "keep entries within the past duration")
	fs.StringVar(&opts.Grep, "grep", "", "substring match on message")
	fs.Var(&opts.Where, "where", "field<op>value expression (repeatable)")
	fs.StringVar(&opts.Output, "output", "color", "color|json|tsv")
	fs.BoolVar(&opts.ShowVersion, "version", false, "print version and exit")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return opts, nil
}

func Run(opts *Options, stdin io.Reader, stdout io.Writer, stderr io.Writer, now time.Time) error {
	if opts.ShowVersion {
		fmt.Fprintln(stdout, "logsift "+Version)
		return nil
	}
	chain, err := buildChain(opts, now)
	if err != nil {
		return err
	}
	w, err := output.New(opts.Output, stdout, opts.Output == "color")
	if err != nil {
		return err
	}

	src := stdin
	if opts.File != "" {
		f, err := os.Open(opts.File)
		if err != nil {
			return err
		}
		defer f.Close()
		src = f
	}

	scanner := bufio.NewScanner(src)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		entry, perr := parser.ParseLine(scanner.Text())
		if perr != nil {
			fmt.Fprintf(stderr, "skip: %v\n", perr)
			continue
		}
		if entry == nil {
			continue
		}
		if !chain.Keep(entry) {
			continue
		}
		if err := w.Write(entry); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func buildChain(opts *Options, now time.Time) (filter.Chain, error) {
	var chain filter.Chain
	if opts.Level != "" {
		chain = append(chain, filter.NewLevelFilter(opts.Level))
	}
	if opts.Since != "" {
		d, err := parser.ParseSince(opts.Since)
		if err != nil {
			return nil, fmt.Errorf("--since: %w", err)
		}
		chain = append(chain, &filter.SinceFilter{Cutoff: now.Add(-d)})
	}
	if opts.Grep != "" {
		chain = append(chain, &filter.GrepFilter{Needle: opts.Grep})
	}
	for _, raw := range opts.Where {
		ex, err := filter.ParseExpr(raw)
		if err != nil {
			return nil, fmt.Errorf("--where: %w", err)
		}
		chain = append(chain, ex)
	}
	return chain, nil
}
