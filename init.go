package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func initCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "init",
		Short: "create a new mder folder",
		Run: func(cmd *cobra.Command, args []string) {
			if err := cloneTemplate(name); err != nil {
				panic(err)
			}
			_, _ = fmt.Fprintf(os.Stdout, "create folder %s success.\n", name)
		},
	}

	cmd.Flags().StringVar(&name, "name", "mder", "Name of the folder to create")
	return cmd
}
