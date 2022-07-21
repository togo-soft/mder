package main

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "mder",
		Short: "mder is a very fast static site generator",
	}
)

func init() {
	// create a new mder folder
	rootCmd.AddCommand(initCmd())
}

func main() {
	rootCmd.Execute()
}
