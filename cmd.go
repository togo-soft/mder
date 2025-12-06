package main

import (
	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"go.abhg.dev/goldmark/mermaid"
)

var (
	rootCmd = &cobra.Command{
		Use:   "mder",
		Short: "mder is a very fast static site generator",
	}

	markdown = goldmark.New(
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithExtensions(extension.GFM, meta.Meta, emoji.Emoji, &mermaid.Extender{}),
	)
)

func init() {
	// create a new site folder
	rootCmd.AddCommand(initCmd())
	// generate static website
	rootCmd.AddCommand(generateCmd())
	// new post or page
	rootCmd.AddCommand(newCmd())
	// run serve locally
	rootCmd.AddCommand(serveCmd())
	// auto update
	rootCmd.AddCommand(updateCmd())
	// deploy project
	rootCmd.AddCommand(deployCmd())
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		logger.Error("execute command failed", "reason", err)
	}
}
