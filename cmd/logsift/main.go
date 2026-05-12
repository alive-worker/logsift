package main

import (
	"fmt"
	"os"
	"time"

	"github.com/alive-worker/logsift/internal/cli"
)

func main() {
	opts, err := cli.ParseArgs(os.Args[1:], os.Stderr)
	if err != nil {
		os.Exit(2)
	}
	if err := cli.Run(opts, os.Stdin, os.Stdout, os.Stderr, time.Now()); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
