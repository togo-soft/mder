package main

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/radovskyb/watcher"
	"github.com/spf13/cobra"
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

func serveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "serve",
		Short:   "run a serve locally",
		Aliases: []string{"s"},
		Run: func(cmd *cobra.Command, args []string) {
			w := watcher.New()
			w.SetMaxEvents(1)
			w.FilterOps(watcher.Rename, watcher.Move, watcher.Write, watcher.Create)
			if err := w.AddRecursive("./data"); err != nil {
				logger.Error("watch directory data failed", "reason", err)
				return
			}
			if err := w.AddRecursive("./posts"); err != nil {
				logger.Error("watch directory posts failed", "reason", err)
				return
			}
			if err := w.AddRecursive("./pages"); err != nil {
				logger.Error("watch directory pages failed", "reason", err)
				return
			}
			if err := w.AddRecursive("./themes"); err != nil {
				logger.Error("watch directory pages failed", "reason", err)
				return
			}
			if err := w.Add("config.yaml"); err != nil {
				logger.Error("watch file failed", "reason", err)
				return
			}
			// start http server
			r := chi.NewRouter()
			workDir, _ := os.Getwd()
			filesDir := http.Dir(filepath.Join(workDir, "dist"))
			FileServer(r, "/", filesDir)

			server := http.Server{
				Addr:    ":8666",
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
						// 文件变更 更新文件
						if err := generateCmd().Execute(); err != nil {
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
			// 先生成一次
			if err := generateCmd().Execute(); err != nil {
				logger.Error("generate website failed", "reason", err)
				return
			}
			// 3秒一次
			logger.Info("http://127.0.0.1:8666")
			if err := w.Start(time.Second * 3); err != nil {
				logger.Error("watch file failed", "reason", err)
				return
			}
		},
	}
	return cmd
}
