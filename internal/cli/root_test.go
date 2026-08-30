package cli

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

// Credentials the built-in patterns locate, enough of them to say the patterns
// are wired up. mask-go is what holds each pattern to its grammar.
const (
	awsKey      = "AKIA0123456789ABCDEF"
	githubToken = "ghp_0123456789abcdefghijklmnopqrstuvwxyz"
	jwt         = "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.abc123"
)

// run executes a fresh command over stdin and reports what it wrote where.
func run(t *testing.T, stdin io.Reader, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	var out, errOut bytes.Buffer
	cmd := NewRootCommand()
	cmd.SetIn(stdin)
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	// A nil argument list is cobra's signal to read os.Args instead, which
	// would hand this command the flags `go test` was run with.
	if args == nil {
		args = []string{}
	}
	cmd.SetArgs(args)

	err = cmd.Execute()
	return out.String(), errOut.String(), err
}

// redact returns what the command makes of src, and fails the test if it makes an
// error of it instead.
func redact(t *testing.T, src string, args ...string) string {
	t.Helper()

	stdout, _, err := run(t, strings.NewReader(src), args...)
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}
	return stdout
}

func TestRootCommandMasks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "empty input",
			src:  "",
			want: "",
		},
		{
			name: "text holding no credential is passed through",
			src:  "the quick brown fox\njumps over the lazy dog\n",
			want: "the quick brown fox\njumps over the lazy dog\n",
		},
		{
			name: "a github token",
			src:  "token=" + githubToken + "\n",
			want: "token=" + strings.Repeat("*", len(githubToken)) + "\n",
		},
		{
			name: "an aws access key id",
			src:  "AWS_ACCESS_KEY_ID=" + awsKey + "\n",
			want: "AWS_ACCESS_KEY_ID=" + strings.Repeat("*", len(awsKey)) + "\n",
		},
		{
			name: "a json web token",
			src:  "Authorization: Bearer " + jwt + "\n",
			want: "Authorization: Bearer " + strings.Repeat("*", len(jwt)) + "\n",
		},
		{
			name: "two credentials on one line",
			src:  awsKey + " and " + githubToken + "\n",
			want: strings.Repeat("*", len(awsKey)) + " and " + strings.Repeat("*", len(githubToken)) + "\n",
		},
		{
			name: "a credential and nothing else",
			src:  githubToken,
			want: strings.Repeat("*", len(githubToken)),
		},
		{
			name: "the last line is not given a newline it did not have",
			src:  "token=" + githubToken,
			want: "token=" + strings.Repeat("*", len(githubToken)),
		},
		{
			name: "carriage returns survive",
			src:  "token=" + githubToken + "\r\nnext\r\n",
			want: "token=" + strings.Repeat("*", len(githubToken)) + "\r\nnext\r\n",
		},
		{
			name: "blank lines survive",
			src:  "\n\n" + githubToken + "\n\n",
			want: "\n\n" + strings.Repeat("*", len(githubToken)) + "\n\n",
		},
		{
			name: "bytes that are not utf-8 are passed through",
			src:  "\x00\xff\xfe\x80binary\x00\n",
			want: "\x00\xff\xfe\x80binary\x00\n",
		},
		{
			name: "a credential written against bytes that are not utf-8",
			src:  "\xff" + githubToken + "\xff",
			want: "\xff" + strings.Repeat("*", len(githubToken)) + "\xff",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := redact(t, tt.src); got != tt.want {
				t.Errorf("output = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRootCommandFlags(t *testing.T) {
	t.Parallel()

	src := "token=" + githubToken + " end\n"
	n := len(githubToken)

	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "no flags fill with asterisks",
			args: nil,
			want: "token=" + strings.Repeat("*", n) + " end\n",
		},
		{
			name: "fill takes another character",
			args: []string{"--fill", "#"},
			want: "token=" + strings.Repeat("#", n) + " end\n",
		},
		{
			// The fill stands for as many characters as the credential held,
			// counted in runes rather than bytes.
			name: "fill takes a multibyte character",
			args: []string{"--fill", "█"},
			want: "token=" + strings.Repeat("█", n) + " end\n",
		},
		{
			name: "replace substitutes a fixed string",
			args: []string{"--replace", "[REDACTED]"},
			want: "token=[REDACTED] end\n",
		},
		{
			// An empty replacement is a replacement, not the flag left alone.
			name: "replace with an empty string removes the value",
			args: []string{"--replace", ""},
			want: "token= end\n",
		},
		{
			name: "flags joined by an equals sign",
			args: []string{"--fill=-"},
			want: "token=" + strings.Repeat("-", n) + " end\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := redact(t, src, tt.args...); got != tt.want {
				t.Errorf("output = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestRootCommandRejects covers the arguments the command turns down. Each is
// turned down before anything is read, so a run that failed halfway never
// releases credentials the flags had not settled how to redact.
func TestRootCommandRejects(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "a positional argument",
			args:    []string{"file.txt"},
			wantErr: `blot reads standard input and takes no arguments: try "blot < file.txt"`,
		},
		{
			// The suggestion is written from the first argument.
			name:    "several positional arguments",
			args:    []string{"one.txt", "two.txt"},
			wantErr: `blot reads standard input and takes no arguments: try "blot < one.txt"`,
		},
		{
			name:    "a positional argument alongside a flag",
			args:    []string{"--fill", "#", "file.txt"},
			wantErr: `blot reads standard input and takes no arguments: try "blot < file.txt"`,
		},
		{
			name:    "an unknown flag",
			args:    []string{"--nope"},
			wantErr: "unknown flag: --nope",
		},
		{
			name:    "fill given more than one character",
			args:    []string{"--fill", "ab"},
			wantErr: `--fill must be a single character: "ab"`,
		},
		{
			name:    "fill given nothing",
			args:    []string{"--fill", ""},
			wantErr: `--fill must be a single character: ""`,
		},
		{
			name:    "fill and replace together",
			args:    []string{"--fill", "#", "--replace", "X"},
			wantErr: "--fill and --replace cannot be used together",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			src := "token=" + githubToken + "\n"
			stdout, stderr, err := run(t, strings.NewReader(src), tt.args...)

			if err == nil {
				t.Fatalf("Execute() error = nil, want %q", tt.wantErr)
			}
			if err.Error() != tt.wantErr {
				t.Errorf("Execute() error = %q, want %q", err.Error(), tt.wantErr)
			}
			if stdout != "" {
				t.Errorf("stdout = %q, want empty", stdout)
			}
			if !strings.Contains(stderr, tt.wantErr) {
				t.Errorf("stderr = %q, want it to hold %q", stderr, tt.wantErr)
			}
			// SilenceUsage: a mistake in the arguments is not a request for
			// the usage text.
			if strings.Contains(stderr, "Usage:") {
				t.Errorf("stderr = %q, want no usage text", stderr)
			}
		})
	}
}

// TestRootCommandChunking holds the command to the same output however the
// input is broken up. A credential split across two reads is in neither of
// them, so this is what says the stream holds text back rather than masking
// each read as it arrives.
func TestRootCommandChunking(t *testing.T) {
	t.Parallel()

	// readSize is how much mask-go's Reader asks for at a time.
	const readSize = 4 << 10

	inputs := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "a short line",
			src:  "token=" + githubToken + "\n",
			want: "token=" + strings.Repeat("*", len(githubToken)) + "\n",
		},
		{
			name: "a credential laid across the read boundary",
			src:  strings.Repeat("-", readSize-len(githubToken)/2) + githubToken + "\n",
			want: strings.Repeat("-", readSize-len(githubToken)/2) + strings.Repeat("*", len(githubToken)) + "\n",
		},
		{
			name: "a credential at the very end of a long input",
			src:  strings.Repeat("-", 3*readSize) + githubToken,
			want: strings.Repeat("-", 3*readSize) + strings.Repeat("*", len(githubToken)),
		},
		{
			name: "many credentials over many reads",
			src:  strings.Repeat(githubToken+"\n"+strings.Repeat("-", 100)+"\n", 200),
			want: strings.Repeat(strings.Repeat("*", len(githubToken))+"\n"+strings.Repeat("-", 100)+"\n", 200),
		},
	}

	chunkings := []struct {
		name string
		wrap func(io.Reader) io.Reader
	}{
		{name: "whole", wrap: func(r io.Reader) io.Reader { return r }},
		{name: "one byte at a time", wrap: func(r io.Reader) io.Reader { return iotest.OneByteReader(r) }},
		{name: "half at a time", wrap: func(r io.Reader) io.Reader { return iotest.HalfReader(r) }},
		{name: "last read carrying eof", wrap: func(r io.Reader) io.Reader { return iotest.DataErrReader(r) }},
	}

	for _, in := range inputs {
		for _, c := range chunkings {
			t.Run(in.name+"/"+c.name, func(t *testing.T) {
				t.Parallel()

				stdout, _, err := run(t, c.wrap(strings.NewReader(in.src)))
				if err != nil {
					t.Fatalf("Execute() error = %v, want nil", err)
				}
				if stdout != in.want {
					t.Errorf("output length = %d, want %d; output = %.120q..., want %.120q...",
						len(stdout), len(in.want), stdout, in.want)
				}
			})
		}
	}
}

func TestRootCommandReadError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("read failed")

	t.Run("the error is reported", func(t *testing.T) {
		t.Parallel()

		_, _, err := run(t, iotest.ErrReader(wantErr))
		if !errors.Is(err, wantErr) {
			t.Errorf("Execute() error = %v, want %v", err, wantErr)
		}
	})

	t.Run("text settled before the error is still written", func(t *testing.T) {
		t.Parallel()

		src := io.MultiReader(
			strings.NewReader("token="+githubToken+"\nkept\n"),
			iotest.ErrReader(wantErr),
		)
		stdout, _, err := run(t, src)

		if !errors.Is(err, wantErr) {
			t.Errorf("Execute() error = %v, want %v", err, wantErr)
		}
		want := "token=" + strings.Repeat("*", len(githubToken)) + "\nkept\n"
		if stdout != want {
			t.Errorf("output = %q, want %q", stdout, want)
		}
	})
}

// errWriter fails every write, standing for a closed pipe downstream.
type errWriter struct{ err error }

func (w errWriter) Write([]byte) (int, error) { return 0, w.err }

func TestRootCommandWriteError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("write failed")

	cmd := NewRootCommand()
	cmd.SetIn(strings.NewReader("token=" + githubToken + "\n"))
	cmd.SetOut(errWriter{err: wantErr})
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); !errors.Is(err, wantErr) {
		t.Errorf("Execute() error = %v, want %v", err, wantErr)
	}
}

func TestRootCommandHelp(t *testing.T) {
	t.Parallel()

	stdout, _, err := run(t, strings.NewReader(""), "--help")
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}
	for _, want := range []string{"blot", "--fill", "--replace"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("help output = %q, want it to hold %q", stdout, want)
		}
	}
}

// TestNewRootCommandIsIndependent holds NewRootCommand to returning a command
// that carries nothing over from another. A command shared between runs, or
// flags stored outside it, would have a run after one that gave --replace
// redact by replacement without having been asked to.
func TestNewRootCommandIsIndependent(t *testing.T) {
	t.Parallel()

	src := "token=" + githubToken + "\n"
	filled := "token=" + strings.Repeat("*", len(githubToken)) + "\n"

	t.Run("after a run that set replace", func(t *testing.T) {
		t.Parallel()

		if got := redact(t, src, "--replace", "X"); got != "token=X\n" {
			t.Fatalf("first output = %q, want %q", got, "token=X\n")
		}
		if got := redact(t, src); got != filled {
			t.Errorf("second output = %q, want %q", got, filled)
		}
	})

	t.Run("after a run that set fill", func(t *testing.T) {
		t.Parallel()

		want := "token=" + strings.Repeat("#", len(githubToken)) + "\n"
		if got := redact(t, src, "--fill", "#"); got != want {
			t.Fatalf("first output = %q, want %q", got, want)
		}
		if got := redact(t, src); got != filled {
			t.Errorf("second output = %q, want %q", got, filled)
		}
	})

	t.Run("after a run that was rejected", func(t *testing.T) {
		t.Parallel()

		if _, _, err := run(t, strings.NewReader(src), "--fill", "#", "--replace", "X"); err == nil {
			t.Fatal("first Execute() error = nil, want the flags to be rejected")
		}
		if got := redact(t, src); got != filled {
			t.Errorf("second output = %q, want %q", got, filled)
		}
	})

	t.Run("two commands do not share flag storage", func(t *testing.T) {
		t.Parallel()

		a, b := NewRootCommand(), NewRootCommand()
		if err := a.Flags().Set("fill", "#"); err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		if got, _ := b.Flags().GetString("fill"); got != "*" {
			t.Errorf("second command's fill = %q, want %q", got, "*")
		}
		if b.Flags().Changed("fill") {
			t.Error("second command reports fill as given, want it left alone")
		}
	})
}
