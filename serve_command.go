package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/radovskyb/watcher"
	"github.com/spf13/cobra"
)

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
				logger.Errorf("watch directory data failed: %v", err)
				return
			}
			if err := w.AddRecursive("./posts"); err != nil {
				logger.Errorf("watch directory posts failed: %v", err)
				return
			}
			if err := w.AddRecursive("./pages"); err != nil {
				logger.Errorf("watch directory pages failed: %v", err)
				return
			}
			if err := w.AddRecursive("./themes"); err != nil {
				logger.Errorf("watch directory pages failed: %v", err)
				return
			}
			if err := w.Add("config.yaml"); err != nil {
				logger.Errorf("watch file failed: %v", err)
				return
			}
			// start http server
			gin.SetMode(gin.ReleaseMode)
			router := gin.New()
			router.Static("/", "./dist")
			server := http.Server{
				Addr:    ":8666",
				Handler: router,
			}
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Errorf("listen server failed: %v", err)
				}
			}()

			go func() {
				for {
					select {
					case event := <-w.Event:
						// 文件变更 更新文件
						if err := generateCmd().Execute(); err != nil {
							logger.Errorf("re generate website failed: %v", err)
							return
						}
						logger.Infof("file change: %v", event.String())
					case err := <-w.Error:
						logger.Errorf("watch file failed: %v", err)
					case <-w.Closed:
						return
					}
				}
			}()
			// 先生成一次
			if err := generateCmd().Execute(); err != nil {
				logger.Errorf("generate website failed: %v", err)
				return
			}
			// 3秒一次
			logger.Info("http://127.0.0.1:8666")
			if err := w.Start(time.Second * 3); err != nil {
				logger.Errorf("watch file failed: %v", err)
				return
			}
		},
	}
	return cmd
}
