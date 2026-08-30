package cli

import (
	"errors"
	"fmt"
	"unicode/utf8"

	mask "github.com/koki-develop/mask-go"
)

// newRedactor builds the redactor the flags ask for.
//
// fillSet and replaceSet say whether each flag was given on the command line,
// which the values alone cannot: --replace with an empty string asks for the
// value to be dropped, and that is what --replace left alone also reads as.
func newRedactor(fill, replace string, fillSet, replaceSet bool) (mask.Redactor, error) {
	if fillSet && replaceSet {
		return nil, errors.New("--fill and --replace cannot be used together")
	}
	if replaceSet {
		return mask.Fixed(replace), nil
	}
	r, size := utf8.DecodeRuneInString(fill)
	if size != len(fill) || (r == utf8.RuneError && size <= 1) {
		return nil, fmt.Errorf("--fill must be a single character: %q", fill)
	}
	return mask.Fill(r), nil
}
