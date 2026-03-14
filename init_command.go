package main

import (
	"embed"
	"fmt"
	"regexp"

	"github.com/spf13/cobra"
)

//go:embed all:template
var frameworkTpl embed.FS

func initCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "init",
		Short: "create a new mder project",
		Run: func(cmd *cobra.Command, args []string) {
			if name == "" && len(args) == 0 {
				logger.Error("folder name empty")
				return
			}
			if name == "" && len(args) != 0 {
				name = args[0]
			}
			var rule = fmt.Sprintf(`[A-Za-z0-9_]{%d}`, len([]rune(name)))
			var reg = regexp.MustCompilePOSIX(rule)
			if !reg.MatchString(name) {
				logger.Error("folder name rule must be: " + rule)
				return
			}
			if err := cloneTemplate(name); err != nil {
				logger.Error("clone template repository failed", "reason", err)
				return
			}
			logger.Info("create folder success", "folder", name)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "name of the folder to create")
	return cmd
}
