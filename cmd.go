package main

import (
	"context"
	"errors"
	"flag"
	"os"

	"github.com/urfave/cli/v3"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"go.abhg.dev/goldmark/mermaid"
)

const (
	exitOK      = 0
	exitFailure = 1
	exitUsage   = 2
)

var (
	markdown = goldmark.New(
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithExtensions(extension.GFM, meta.Meta, emoji.Emoji, &mermaid.Extender{}),
	)
)

func newApp() *cli.Command {
	return &cli.Command{
		Name:  "mder",
		Usage: "mder is a very fast static site generator",
		Commands: []*cli.Command{
			initCmd(),
			generateCmd(),
			newCmd(),
			serveCmd(),
			updateCmd(),
		},
	}
}

func runMain(args []string) error {
	return newApp().Run(context.Background(), args)
}

func exitCode(err error) int {
	if err == nil {
		return exitOK
	}
	var exitCoder cli.ExitCoder
	if errors.As(err, &exitCoder) {
		return exitCoder.ExitCode()
	}
	if errors.Is(err, flag.ErrHelp) {
		return exitUsage
	}
	return exitFailure
}

func main() {
	if err := runMain(os.Args); err != nil {
		logger.Error("execute command failed", "reason", err)
		os.Exit(exitCode(err))
	}
}
