package cli

import (
	"io"

	mask "github.com/koki-develop/mask-go"
	"github.com/spf13/cobra"
)

// NewRootCommand returns a fresh blot command, with flags of its own.
//
// Each call is independent: cobra records on the command whether a flag was
// given, so a command run twice would read the flags of the first run in the
// second.
func NewRootCommand() *cobra.Command {
	var (
		fillFlag    string
		replaceFlag string
	)

	cmd := &cobra.Command{
		Use:          "blot",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := newRedactor(
				fillFlag, replaceFlag,
				cmd.Flags().Changed("fill"), cmd.Flags().Changed("replace"),
			)
			if err != nil {
				return err
			}
			m := mask.New(
				mask.WithPatterns(mask.AllBuiltinPatterns()...),
				mask.WithRedactor(r),
			)
			_, err = io.Copy(cmd.OutOrStdout(), mask.NewReader(cmd.InOrStdin(), m))
			return err
		},
	}

	cmd.Flags().StringVar(&fillFlag, "fill", "*", "character to repeat over each masked value, preserving its length")
	cmd.Flags().StringVar(&replaceFlag, "replace", "", "string to replace each masked value with, discarding its length")

	return cmd
}

func Execute() error {
	return NewRootCommand().Execute()
}
