package main

import (
	"log/slog"
	"os"

	"github.com/golang-cz/devslog"
)

var logger *slog.Logger

func init() {
	slogOpts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}
	opts := &devslog.Options{
		HandlerOptions:    slogOpts,
		MaxSlicePrintSize: 4,
		SortKeys:          true,
		TimeFormat:        "[04:05]",
		NewLineAfterLog:   true,
		DebugColor:        devslog.Magenta,
		StringerFormatter: true,
	}

	logger = slog.New(devslog.NewHandler(os.Stdout, opts))
	slog.SetDefault(logger)
}
