package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "blot",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}
