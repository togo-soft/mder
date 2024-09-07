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
		PreRun: func(cmd *cobra.Command, args []string) {
			if path != "" {
				BaseDir = strings.TrimSuffix(path, "/")
			}

			if err := outter.loadConfig(); err != nil {
				logger.Fatalf("load config failed: %v", err)
			}
			config := outter.Config.Deploy
			switch config.Type {
			case UpyunDeploy:
				if !isCommandExist("upx") {
					if err := goInstall("github.com/upyun/upx/cmd/upx"); err != nil {
						logger.Errorf("install upx failed: %+v", err)
						return
					}
				}
			case GitDeploy:
			}

			// generate source
			if err := generateCmd().Execute(); err != nil {
				logger.Errorf("generate project failed: %v", err)
				return
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			config := outter.Config.Deploy
			switch config.Type {
			case UpyunDeploy:
				if config.UpyunAuth == "" {
					logger.Errorf("please config upyun auth string: upx auth [bucket] [operator] [password]")
					return
				}
				if err := uploadToUpyun(config.UpyunAuth, path); err != nil {
					logger.Errorf("upload dist to upyun failed: %v", err)
					return
				}
				logger.Infof("deploy to upyun success")
			}
		},
	}
	cmd.Flags().StringVar(&path, "path", ".", "mder project path")
	return cmd
}
