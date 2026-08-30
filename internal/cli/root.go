package cli

import (
	"fmt"
	"io"

	mask "github.com/koki-develop/mask-go"
	"github.com/spf13/cobra"
)

// NewRootCommand returns a fresh blot command, with flags of its own.
//
// Each call is independent: cobra records on the command whether a flag was
// given, so a command run twice would read the flags of the first run in the
// second.
func NewRootCommand(version string) *cobra.Command {
	var (
		fillFlag    string
		replaceFlag string
	)

	cmd := &cobra.Command{
		Use:          "blot",
		Version:      version,
		Args:         noArgs,
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

// noArgs turns down whatever blot is given on the command line, since it reads
// standard input and nothing else.
//
// cobra.NoArgs reports an unknown command, which points away from the mistake:
// blot has no subcommands, so an argument here is a file someone meant to have
// read rather than a command they got wrong.
func noArgs(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}
	return fmt.Errorf("%s reads standard input and takes no arguments: try %q",
		cmd.CommandPath(), cmd.CommandPath()+" < "+args[0])
}

// Execute runs blot. stamped is the version goreleaser wrote into the binary,
// empty in a build it did not make.
func Execute(stamped string) error {
	return NewRootCommand(resolveVersion(stamped, moduleVersion())).Execute()
}
