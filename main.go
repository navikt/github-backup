package main

import (
	"context"
	"fmt"
	"os"

	"github.com/navikt/github-backup/internal/githubbackup"
)

func main() {
	if err := githubbackup.Run(context.Background()); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
