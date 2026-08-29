package cli

import (
	"errors"
	"fmt"
	"io"
	"unicode/utf8"

	mask "github.com/koki-develop/mask-go"
	"github.com/spf13/cobra"
)

var (
	fillFlag    string
	replaceFlag string
)

var rootCmd = &cobra.Command{
	Use:          "blot",
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := redactor(cmd)
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

func redactor(cmd *cobra.Command) (mask.Redactor, error) {
	if cmd.Flags().Changed("fill") && cmd.Flags().Changed("replace") {
		return nil, errors.New("--fill and --replace cannot be used together")
	}
	if cmd.Flags().Changed("replace") {
		return mask.Fixed(replaceFlag), nil
	}
	r, size := utf8.DecodeRuneInString(fillFlag)
	if size != len(fillFlag) || (r == utf8.RuneError && size <= 1) {
		return nil, fmt.Errorf("--fill must be a single character: %q", fillFlag)
	}
	return mask.Fill(r), nil
}

func init() {
	rootCmd.Flags().StringVar(&fillFlag, "fill", "*", "character to repeat over each masked value, preserving its length")
	rootCmd.Flags().StringVar(&replaceFlag, "replace", "", "string to replace each masked value with, discarding its length")
}

func Execute() error {
	return rootCmd.Execute()
}
