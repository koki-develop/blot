# blot

[![GitHub Release](https://img.shields.io/github/v/release/koki-develop/blot?style=flat-square)](https://github.com/koki-develop/blot/releases/latest)
[![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/koki-develop/blot/ci.yml?style=flat-square&logo=github)](https://github.com/koki-develop/blot/actions/workflows/ci.yml)
[![GitHub License](https://img.shields.io/github/license/koki-develop/blot?style=flat-square)](./LICENSE)

Secret masking filter.

## Installation

### Homebrew

```sh
brew install koki-develop/tap/blot
```

### go install

```sh
go install github.com/koki-develop/blot@latest
```

Prebuilt binaries are also on the [releases page](https://github.com/koki-develop/blot/releases/latest).

## Usage

blot reads standard input, redacts every credential it finds and writes the
result to standard output:

```console
$ echo 'GITHUB_TOKEN=ghp_0123456789abcdefghijklmnopqrstuvwxyz' | blot
GITHUB_TOKEN=****************************************
```

## Flags

`--fill` sets the character repeated over each value, which preserves its
length (default `*`):

```console
$ echo 'GITHUB_TOKEN=ghp_0123456789abcdefghijklmnopqrstuvwxyz' | blot --fill '#'
GITHUB_TOKEN=########################################
```

`--replace` replaces each value with a fixed string, which discards its length:

```console
$ echo 'GITHUB_TOKEN=ghp_0123456789abcdefghijklmnopqrstuvwxyz' | blot --replace '[REDACTED]'
GITHUB_TOKEN=[REDACTED]
```

## Patterns

blot looks for every pattern built into
[mask-go](https://github.com/koki-develop/mask-go#patterns): tokens and keys
from GitHub, GitLab, AWS, Google, Stripe, Slack, OpenAI, Anthropic, Supabase
and others, alongside JSON Web Tokens and PEM private keys.

## License

[MIT](./LICENSE)
