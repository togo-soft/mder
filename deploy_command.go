package main

import (
	"strings"

	"github.com/spf13/cobra"
)

func deployCmd() *cobra.Command {
	var path string
	var outter Outter
	cmd := &cobra.Command{
		Use:     "deploy",
		Aliases: []string{"d"},
		Short:   "deploy project to upyun",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if path != "" {
				BaseDir = strings.TrimSuffix(path, "/")
			}

			if err := outter.loadConfig(); err != nil {
				logger.Error("load config failed", "reason", err)
				return err
			}
			config := outter.Config.Deploy
			switch config.Type {
			case UpyunDeploy:
				if !isCommandExist("upx") {
					if err := goInstall("github.com/upyun/upx/cmd/upx"); err != nil {
						logger.Error("install upx failed", "reason", err)
						return err
					}
				}
			case GitHubDeploy:
			}

			// generate source
			if err := generateCmd().Execute(); err != nil {
				logger.Error("generate project failed", "reason", err)
				return err
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			config := outter.Config.Deploy
			switch config.Type {
			case UpyunDeploy:
				if config.UpyunAuth == "" {
					logger.Error("please config upyun auth string: upx auth [bucket] [operator] [password]")
					return
				}
				if err := uploadToUpyun(config.UpyunAuth, path); err != nil {
					logger.Error("upload dist to upyun failed", "reason", err)
					return
				}
				logger.Info("deploy to upyun success")
			case GitHubDeploy:
			}

		},
	}
	cmd.Flags().StringVar(&path, "path", ".", "mder project path")
	return cmd
}
