# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

blot filters stdin to stdout, redacting credentials. All pattern matching and streaming lives in
[mask-go](https://github.com/koki-develop/mask-go) — this repo only maps flags to a `mask.Redactor`.
Adding or fixing a pattern is a change to mask-go, not here.

## Commands

The toolchain is pinned in `mise.toml`. Run `mise install`, then `mise run bootstrap` for the lefthook
git hooks.

```sh
go test ./...
go test ./internal/cli -run TestNewRedactor
golangci-lint run ./...
goreleaser check && goreleaser release --snapshot --clean  # what CI's build job runs
```

## Gotchas

- `cmd/formula/` prints the Homebrew formula for an already-published version. It is never released
  (`.goreleaser.yaml` sets `main: .`), and its `archives` list must match that file's archive
  `name_template`.
- Tests drive the command through `NewRootCommand(version)` with buffers, so I/O must go through
  `cmd.InOrStdin()` / `cmd.OutOrStdout()`, never `os.Stdin` / `os.Stdout`.
- release-please owns the version and creates a **draft** release; goreleaser uploads the assets and
  publishes that draft. Never bump a version by hand.

## Conventions

- Conventional Commits are load-bearing: they drive the release.
- Comments explain *why* a non-obvious choice was made, never what the code already says.
- Tests are table-driven with `t.Parallel()` on both the parent test and each subtest.
