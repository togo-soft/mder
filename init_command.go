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
				logger.Errorf("folder name empty")
				return
			}
			if name == "" && len(args) != 0 {
				name = args[0]
			}
			var rule = fmt.Sprintf(`[A-Za-z0-9_]{%d}`, len([]rune(name)))
			var reg = regexp.MustCompilePOSIX(rule)
			if !reg.MatchString(name) {
				logger.Errorf("folder name rule must be: %s", rule)
				return
			}
			if err := cloneTemplate(name); err != nil {
				logger.Errorf("clone template repository failed: %v", err)
				return
			}
			logger.Infof("create folder `%s` success", name)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "name of the folder to create")
	return cmd
}
