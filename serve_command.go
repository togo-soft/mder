package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/radovskyb/watcher"
	"github.com/urfave/cli/v3"
)

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

func serveCmd() *cli.Command {
	return &cli.Command{
		Name:    "serve",
		Usage:   "run a serve locally",
		Aliases: []string{"s"},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "host",
				Value: "",
				Usage: "listen host address (e.g. 0.0.0.0, ::, 127.0.0.1), empty for all interfaces",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runServe(cmd.String("host"))
		},
	}
}

func runServe(host string) error {
	w := watcher.New()
	w.SetMaxEvents(1)
	w.FilterOps(watcher.Rename, watcher.Move, watcher.Write, watcher.Create)
	if err := w.AddRecursive("./data"); err != nil {
		return fmt.Errorf("watch directory data: %w", err)
	}
	if err := w.AddRecursive("./posts"); err != nil {
		return fmt.Errorf("watch directory posts: %w", err)
	}
	if err := w.AddRecursive("./pages"); err != nil {
		return fmt.Errorf("watch directory pages: %w", err)
	}
	if err := w.AddRecursive("./themes"); err != nil {
		return fmt.Errorf("watch directory themes: %w", err)
	}
	if err := w.Add("config.yaml"); err != nil {
		return fmt.Errorf("watch config file: %w", err)
	}

	r := chi.NewRouter()
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "dist"))
	FileServer(r, "/", filesDir)

	addr := net.JoinHostPort(host, "8666")
	server := http.Server{
		Addr:    addr,
		Handler: r,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen server failed", "reason", err)
		}
	}()

	go func() {
		for {
			select {
			case event := <-w.Event:
				if err := runGenerate("."); err != nil {
					logger.Error("re generate website failed", "reason", err)
					return
				}
				logger.Info("file change", "event", event.String())
			case err := <-w.Error:
				logger.Error("watch file failed", "reason", err)
			case <-w.Closed:
				return
			}
		}
	}()

	if err := runGenerate("."); err != nil {
		return fmt.Errorf("generate website: %w", err)
	}
	logger.Info("http://" + addr)
	if err := w.Start(time.Second * 3); err != nil {
		return fmt.Errorf("watch file: %w", err)
	}
	return nil
}
