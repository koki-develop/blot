// Command formula prints the Homebrew formula for a released version of blot.
//
// Each sha256 is computed by downloading the archive, so the release must
// already be published.
package main

import (
	"crypto/sha256"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/template"
	"time"
)

// archives are the release archives the formula installs from, named as
// .goreleaser.yaml's archive name_template renders them.
var archives = []string{
	"Darwin_x86_64",
	"Darwin_arm64",
	"Linux_x86_64",
	"Linux_arm64",
}

var formula = template.Must(template.New("formula").Parse(`class Blot < Formula
  desc "Secret masking filter"
  homepage "https://github.com/koki-develop/blot"
  version "{{ .Version }}"
  license "MIT"

  on_macos do
    on_intel do
      url "https://github.com/koki-develop/blot/releases/download/v#{version}/blot_Darwin_x86_64.tar.gz"
      sha256 "{{ index .SHA256 "Darwin_x86_64" }}"
    end
    on_arm do
      url "https://github.com/koki-develop/blot/releases/download/v#{version}/blot_Darwin_arm64.tar.gz"
      sha256 "{{ index .SHA256 "Darwin_arm64" }}"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/koki-develop/blot/releases/download/v#{version}/blot_Linux_x86_64.tar.gz"
      sha256 "{{ index .SHA256 "Linux_x86_64" }}"
    end
    on_arm do
      url "https://github.com/koki-develop/blot/releases/download/v#{version}/blot_Linux_arm64.tar.gz"
      sha256 "{{ index .SHA256 "Linux_arm64" }}"
    end
  end

  def install
    bin.install "blot"
  end

  test do
    system bin/"blot", "--help"
  end
end
`))

var client = &http.Client{Timeout: 30 * time.Second}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	version := flag.String("version", "", "released version to generate the formula for, without the leading v (e.g. 0.1.0)")
	flag.Parse()
	if *version == "" {
		return errors.New("-version is required")
	}

	sums := make(map[string]string, len(archives))
	for _, archive := range archives {
		url := fmt.Sprintf("https://github.com/koki-develop/blot/releases/download/v%s/blot_%s.tar.gz", *version, archive)
		sum, err := sha256Of(url)
		if err != nil {
			return err
		}
		sums[archive] = sum
	}

	return formula.Execute(os.Stdout, struct {
		Version string
		SHA256  map[string]string
	}{Version: *version, SHA256: sums})
}

func sha256Of(url string) (string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("downloading %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("downloading %s: %s", url, resp.Status)
	}

	h := sha256.New()
	if _, err := io.Copy(h, resp.Body); err != nil {
		return "", fmt.Errorf("reading %s: %w", url, err)
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
