package main

import (
	"fmt"
	"regexp"

	"github.com/spf13/cobra"
)

func initCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "init",
		Short: "create a new mder folder",
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
