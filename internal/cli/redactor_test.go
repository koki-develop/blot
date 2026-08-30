package cli

import (
	"testing"

	mask "github.com/koki-develop/mask-go"
)

// TestNewRedactor covers which redactor the flags select and what it does to a
// value. mask.Redactor is an interface with nothing observable on it but
// Redact, so every case states a value and what it must come back as.
func TestNewRedactor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		fill       string
		replace    string
		fillSet    bool
		replaceSet bool
		value      string
		want       string
		wantErr    string // the exact error expected, or empty for none
	}{
		// --fill, unset and set. Fill repeats its character once per rune of
		// the value, so three runes are redacted to three characters however
		// many bytes they were written in.
		{
			name:  "default fills with an asterisk per rune",
			fill:  "*",
			value: "aあb",
			want:  "***",
		},
		{
			name:    "fill takes an ascii character",
			fill:    "#",
			fillSet: true,
			value:   "abc",
			want:    "###",
		},
		{
			name:    "fill takes a multibyte rune",
			fill:    "あ",
			fillSet: true,
			value:   "abc",
			want:    "あああ",
		},
		{
			name:    "fill takes a rune outside the basic plane",
			fill:    "\U0001F512",
			fillSet: true,
			value:   "ab",
			want:    "\U0001F512\U0001F512",
		},
		{
			name:    "fill takes a combining mark",
			fill:    "́",
			fillSet: true,
			value:   "ab",
			want:    "́́",
		},
		{
			// U+FFFD is a character in its own right and three bytes of valid
			// UTF-8, so it is told apart from the single byte DecodeRune
			// reports it for and is accepted.
			name:    "fill takes a literal replacement character",
			fill:    "�",
			fillSet: true,
			value:   "ab",
			want:    "��",
		},
		{
			name:    "fill set to the default value is still the default",
			fill:    "*",
			fillSet: true,
			value:   "abc",
			want:    "***",
		},
		{
			name:  "fill redacts an empty value to nothing",
			fill:  "*",
			value: "",
			want:  "",
		},

		// --fill, rejected. A character is one rune, which is narrower than
		// what reads as one character on screen.
		{
			name:    "fill rejects an empty string",
			fill:    "",
			fillSet: true,
			wantErr: `--fill must be a single character: ""`,
		},
		{
			name:    "fill rejects two characters",
			fill:    "ab",
			fillSet: true,
			wantErr: `--fill must be a single character: "ab"`,
		},
		{
			name:    "fill rejects a regional indicator pair",
			fill:    "\U0001F1EF\U0001F1F5",
			fillSet: true,
			wantErr: `--fill must be a single character: "🇯🇵"`,
		},
		{
			name:    "fill rejects a zero width joiner sequence",
			fill:    "\U0001F468\u200d\U0001F469\u200d\U0001F466",
			fillSet: true,
			wantErr: `--fill must be a single character: "👨\u200d👩\u200d👦"`,
		},
		{
			name:    "fill rejects a lone invalid byte",
			fill:    "\xff",
			fillSet: true,
			wantErr: `--fill must be a single character: "\xff"`,
		},
		{
			name:    "fill rejects an invalid byte in front of a character",
			fill:    "\xffx",
			fillSet: true,
			wantErr: `--fill must be a single character: "\xffx"`,
		},
		{
			name:    "fill rejects a truncated multibyte sequence",
			fill:    "\xe3\x81",
			fillSet: true,
			wantErr: `--fill must be a single character: "\xe3\x81"`,
		},

		// --replace. The flag being given is what selects it: an empty
		// replacement asks for the value to be dropped, which is what the flag
		// left alone also holds.
		{
			name:       "replace substitutes a fixed string",
			replace:    "[REDACTED]",
			replaceSet: true,
			value:      "abc",
			want:       "[REDACTED]",
		},
		{
			name:       "replace ignores the length of the value",
			replace:    "X",
			replaceSet: true,
			value:      "a very long secret value",
			want:       "X",
		},
		{
			name:       "replace with an empty string drops the value",
			replace:    "",
			replaceSet: true,
			value:      "abc",
			want:       "",
		},

		// Both flags.
		{
			name:       "fill and replace together are rejected",
			fill:       "#",
			replace:    "X",
			fillSet:    true,
			replaceSet: true,
			wantErr:    "--fill and --replace cannot be used together",
		},
		{
			name:       "the conflict is reported ahead of an invalid fill",
			fill:       "ab",
			replace:    "X",
			fillSet:    true,
			replaceSet: true,
			wantErr:    "--fill and --replace cannot be used together",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := newRedactor(tt.fill, tt.replace, tt.fillSet, tt.replaceSet)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("newRedactor() error = nil, want %q", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("newRedactor() error = %q, want %q", err.Error(), tt.wantErr)
				}
				if r != nil {
					t.Errorf("newRedactor() redactor = %v, want nil alongside an error", r)
				}
				return
			}

			if err != nil {
				t.Fatalf("newRedactor() error = %v, want nil", err)
			}
			if r == nil {
				t.Fatal("newRedactor() redactor = nil, want non-nil")
			}
			if got := r.Redact(mask.Match{Value: tt.value}); got != tt.want {
				t.Errorf("Redact(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}
