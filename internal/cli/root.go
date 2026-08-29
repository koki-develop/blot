package cli

import (
	"io"

	mask "github.com/koki-develop/mask-go"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "blot",
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		m := mask.New(mask.WithPatterns(mask.AllBuiltinPatterns()...))
		_, err := io.Copy(cmd.OutOrStdout(), mask.NewReader(cmd.InOrStdin(), m))
		return err
	},
}

func Execute() error {
	return rootCmd.Execute()
}
