package cli

import (
	"runtime/debug"
	"strings"
)

// resolveVersion settles what --version reports.
//
// goreleaser stamps the version without a leading v and a module version
// carries one, so the v is trimmed off either and both read alike.
func resolveVersion(stamped, module string) string {
	if v := strings.TrimPrefix(stamped, "v"); v != "" {
		return v
	}
	if v := strings.TrimPrefix(module, "v"); v != "" {
		return v
	}
	return "dev"
}

// moduleVersion is the version of the module the binary was built from, which
// `go install ...@v0.1.0` knows even though nothing was stamped.
//
// "(devel)" is what a build outside a released module reads as, which says no
// more than what resolveVersion falls back to.
func moduleVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "(devel)" {
		return ""
	}
	return info.Main.Version
}
