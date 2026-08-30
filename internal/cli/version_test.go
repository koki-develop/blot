package cli

import "testing"

// TestResolveVersion covers what --version reports for a binary however it was
// built: stamped by goreleaser, installed from the module, or neither.
func TestResolveVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		stamped string
		module  string
		want    string
	}{
		{
			name:    "a stamped version is reported as it is",
			stamped: "0.1.0",
			want:    "0.1.0",
		},
		{
			name:    "a stamped snapshot version is reported as it is",
			stamped: "0.1.1-next-SNAPSHOT-abcdef1",
			want:    "0.1.1-next-SNAPSHOT-abcdef1",
		},
		{
			name:    "a stamped version keeps no leading v",
			stamped: "v0.1.0",
			want:    "0.1.0",
		},
		{
			name:    "the stamp is taken over the module",
			stamped: "0.1.0",
			module:  "v0.2.0",
			want:    "0.1.0",
		},
		{
			// What `go install ...@v0.1.0` builds: nothing stamped, and a
			// module version written the way a tag is.
			name:   "the module version keeps no leading v",
			module: "v0.1.0",
			want:   "0.1.0",
		},
		{
			name:   "a module pseudo-version is reported as it is",
			module: "v0.0.0-20260830023332-ce6e2572ee1f",
			want:   "0.0.0-20260830023332-ce6e2572ee1f",
		},
		{
			name:    "a stamp of nothing but a v falls back to the module",
			stamped: "v",
			module:  "v0.1.0",
			want:    "0.1.0",
		},
		{
			name: "neither is reported as a development build",
			want: "dev",
		},
		{
			name:    "a v on either side is not a version",
			stamped: "v",
			module:  "v",
			want:    "dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := resolveVersion(tt.stamped, tt.module); got != tt.want {
				t.Errorf("resolveVersion(%q, %q) = %q, want %q", tt.stamped, tt.module, got, tt.want)
			}
		})
	}
}

// TestModuleVersion holds moduleVersion to what resolveVersion can make sense
// of. The test binary is a build of this module, so what it reports is up to
// how the test itself was built: any version but "(devel)".
func TestModuleVersion(t *testing.T) {
	t.Parallel()

	if got := moduleVersion(); got == "(devel)" {
		t.Errorf("moduleVersion() = %q, want a version or nothing", got)
	}
}
