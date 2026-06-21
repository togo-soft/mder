package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"regexp"

	"github.com/urfave/cli/v3"
)

//go:embed all:template
var frameworkTpl embed.FS

func initCmd() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "create a new mder project",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name", Usage: "name of the folder to create"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			name := cmd.String("name")
			if name == "" && cmd.NArg() != 0 {
				name = cmd.Args().First()
			}
			return runInit(name)
		},
	}
}

func runInit(name string) error {
	if name == "" {
		return cli.Exit(errors.New("folder name empty"), exitUsage)
	}
	rule := fmt.Sprintf(`[A-Za-z0-9_]{%d}`, len([]rune(name)))
	reg := regexp.MustCompilePOSIX(rule)
	if !reg.MatchString(name) {
		return cli.Exit(fmt.Errorf("folder name rule must be: %s", rule), exitUsage)
	}
	if err := cloneTemplate(name); err != nil {
		return fmt.Errorf("clone template repository: %w", err)
	}
	logger.Info("create folder success", "folder", name)
	return nil
}
