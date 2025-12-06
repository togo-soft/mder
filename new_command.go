package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new",
		Short: "new a post or page",
	}

	cmd.AddCommand(newPostCmd())
	cmd.AddCommand(newPageCmd())

	return cmd
}

func newPostCmd() *cobra.Command {
	var name, catalog string
	cmd := &cobra.Command{
		Use:   "post",
		Short: "new a post",
		Run: func(cmd *cobra.Command, args []string) {
			if name == "" && len(args) != 0 {
				name = args[0]
			}
			var pureName = strings.ReplaceAll(name, ".md", "")
			pureName = strings.ReplaceAll(name, "-", " ")
			// 空格处理
			name = strings.ReplaceAll(name, " ", "-")
			if !strings.HasSuffix(name, ".md") {
				name = name + ".md"
			}
			filename := fmt.Sprintf("posts/%s", name)
			if catalog != "" {
				_name := strings.ReplaceAll(name, "/", "-")
				filename = fmt.Sprintf("posts/%s/%s", catalog, _name)
			}
			// 检测文件夹是否存在
			dir := filepath.Dir(filename)
			if !isExist(dir) {
				if err := mkdir(dir); err != nil {
					logger.Error("make directory failed", "reason", err)
					return
				}
			}
			f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, os.ModePerm)
			if err != nil {
				logger.Error("create file failed", "reason", err)
				return
			}
			defer f.Close()
			var data = fmt.Sprintf("---\ntitle: %s\ndate: %s\ncatagories: %s\ntags:\n---", pureName, time.Now().Format("2006-01-02 15:04:05"), catalog)
			if _, err := f.Write([]byte(data)); err != nil {
				logger.Error("create file failed", "reason", err)
				return
			}
		},
	}
	cmd.Flags().StringVarP(&name, "name", "n", "uname.md", "Name of the post file to create")
	cmd.Flags().StringVarP(&catalog, "catalog", "c", "develop", "Catalog of the post file to create")
	return cmd
}

func newPageCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "page",
		Short: "new a page",
		Run: func(cmd *cobra.Command, args []string) {
			if name == "" && len(args) != 0 {
				name = args[0]
			}
			var pureName = strings.ReplaceAll(name, ".md", "")
			if !strings.HasSuffix(name, ".md") {
				name = name + ".md"
			}
			filename := fmt.Sprintf("pages/%s", name)
			// 检测文件夹是否存在
			dir := filepath.Dir(filename)
			if !isExist(dir) {
				if err := mkdir(dir); err != nil {
					logger.Error("make directory failed", "reason", err)
					return
				}
			}
			f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, os.ModePerm)
			if err != nil {
				logger.Error("create file failed", "reason", err)
				return
			}
			defer f.Close()
			var data = fmt.Sprintf("---\ntitle: %s\ndate: %s\n---", pureName, time.Now().Format("2006-01-02 15:04:05"))
			if _, err := f.Write([]byte(data)); err != nil {
				logger.Error("create file failed", "reason", err)
				return
			}
		},
	}
	cmd.Flags().StringVar(&name, "name", "uname.md", "Name of the page file to create")
	return cmd
}
