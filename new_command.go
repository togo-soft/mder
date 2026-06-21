package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
)

func newCmd() *cli.Command {
	return &cli.Command{
		Name:  "new",
		Usage: "new a post or page",
		Commands: []*cli.Command{
			newPostCmd(),
			newPageCmd(),
		},
	}
}

func newPostCmd() *cli.Command {
	return &cli.Command{
		Name:  "post",
		Usage: "new a post",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Value: "uname.md", Usage: "Name of the post file to create"},
			&cli.StringFlag{Name: "catalog", Aliases: []string{"c"}, Value: "develop", Usage: "Catalog of the post file to create"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			name := cmd.String("name")
			if name == "" && cmd.NArg() != 0 {
				name = cmd.Args().First()
			}
			return runNewPost(name, cmd.String("catalog"))
		},
	}
}

func newPageCmd() *cli.Command {
	return &cli.Command{
		Name:  "page",
		Usage: "new a page",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name", Value: "uname.md", Usage: "Name of the page file to create"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			name := cmd.String("name")
			if name == "" && cmd.NArg() != 0 {
				name = cmd.Args().First()
			}
			return runNewPage(name)
		},
	}
}

func runNewPost(name, catalog string) error {
	pureName := strings.ReplaceAll(name, ".md", "")
	pureName = strings.ReplaceAll(pureName, "-", " ")
	name = strings.ReplaceAll(name, " ", "-")
	if !strings.HasSuffix(name, ".md") {
		name += ".md"
	}
	filename := fmt.Sprintf("posts/%s", name)
	if catalog != "" {
		sanitizedName := strings.ReplaceAll(name, "/", "-")
		filename = fmt.Sprintf("posts/%s/%s", catalog, sanitizedName)
	}
	if err := ensureParentDir(filename); err != nil {
		return err
	}
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return fmt.Errorf("create post file: %w", err)
	}
	defer f.Close()
	data := fmt.Sprintf("---\ntitle: %s\ndate: %s\ncategories: %s\ntags:\n---", pureName, time.Now().Format("2006-01-02 15:04:05"), catalog)
	if _, err := f.Write([]byte(data)); err != nil {
		return fmt.Errorf("write post file: %w", err)
	}
	return nil
}

func runNewPage(name string) error {
	pureName := strings.ReplaceAll(name, ".md", "")
	if !strings.HasSuffix(name, ".md") {
		name += ".md"
	}
	filename := fmt.Sprintf("pages/%s", name)
	if err := ensureParentDir(filename); err != nil {
		return err
	}
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return fmt.Errorf("create page file: %w", err)
	}
	defer f.Close()
	data := fmt.Sprintf("---\ntitle: %s\ndate: %s\n---", pureName, time.Now().Format("2006-01-02 15:04:05"))
	if _, err := f.Write([]byte(data)); err != nil {
		return fmt.Errorf("write page file: %w", err)
	}
	return nil
}

func ensureParentDir(filename string) error {
	dir := filepath.Dir(filename)
	if isExist(dir) {
		return nil
	}
	if err := mkdir(dir); err != nil {
		return fmt.Errorf("make directory: %w", err)
	}
	return nil
}
