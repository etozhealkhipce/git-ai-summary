# git-ai-summary

CLI tool: reads **git** history for a time range, sends it to **OpenAI**, **Anthropic**, or any **OpenAI-compatible** API, and prints an AI-generated summary (for demos or reports).

## Requirements

- **Prebuilt binary:** `git` on your `PATH` (and `curl` if you use the install script).
- **From source:** [Go](https://go.dev/dl/) 1.22+ and `git`.

## Install

Binaries for **macOS**, **Linux**, and **Windows** are on [Releases](https://github.com/etozhealkhipce/git-ai-summary/releases).

**macOS / Linux** (latest):

```bash
curl -sSL https://raw.githubusercontent.com/etozhealkhipce/git-ai-summary/main/install.sh | sh
```

Pin a version (`0.1.0`, no `v` prefix):

```bash
GIT_AI_SUMMARY_VERSION=0.1.0 curl -sSL https://raw.githubusercontent.com/etozhealkhipce/git-ai-summary/main/install.sh | sh
```

Optional: `GIT_AI_SUMMARY_INSTALL_DIR` (default: `$HOME/.local/bin`). Safer: clone the repo and run `./install.sh` after reading the script.

**Windows (PowerShell):**

```powershell
irm https://raw.githubusercontent.com/etozhealkhipce/git-ai-summary/main/install.ps1 | iex
```

**Go install:**

```bash
go install github.com/etozhealkhipce/git-ai-summary/cmd/git-ai-summary@latest
```

Put `$GOBIN` or `$GOPATH/bin` on your `PATH`.

## API keys

Interactive setup (writes env vars to your shell profile or a PowerShell file):

```bash
git-ai-summary setup
```

Or set the keys yourself (see table below).

## Environment

| Variable | Purpose |
|----------|---------|
| `OPENAI_API_KEY` | OpenAI and OpenAI-compatible providers |
| `ANTHROPIC_API_KEY` | Anthropic |
| `GIT_AI_SUMMARY_PROVIDER` | `openai`, `anthropic`, or `openai-compatible` |
| `GIT_AI_SUMMARY_BASE_URL` | Base URL for `openai-compatible` (needs `/v1` style path if your API uses it) |
| `GIT_AI_SUMMARY_MODEL` | Model id override |

CLI flags override env vars.

## Usage

```text
git-ai-summary [options]
git-ai-summary setup
```

Common options:

| Flag | Notes |
|------|--------|
| `-repo` | Git repo path (default: current directory) |
| `-since` | Passed to `git log --since` (default: `7 days ago`) |
| `-provider` | `openai` (default), `anthropic`, or `openai-compatible` |
| `-format` | `pretty`, `tsv`, `csv`, `md`, `json` — default is **`pretty` in a terminal**, **`tsv`** when stdout is not a TTY or with `-o` (unless you set `-format`) |
| `-o` | Write output to a file |
| `-dry-run` | Print the git bundle only; no API call |

`openai-compatible` needs `-base-url` or `GIT_AI_SUMMARY_BASE_URL`.

Examples:

```bash
git-ai-summary -dry-run
export OPENAI_API_KEY=...
git-ai-summary -o summary.tsv
export ANTHROPIC_API_KEY=...
git-ai-summary -provider anthropic -format md -o summary.md
```

While waiting for the API, the tool shows a short spinner on stderr (TTY only). Set `NO_COLOR` to disable ANSI colors.

## Releases (maintainers)

[Actions](https://github.com/etozhealkhipce/git-ai-summary/actions) → workflow **release** → **Run workflow**. Pick the branch (usually `main`) and bump type: **auto** (from commits since last tag), or **patch** / **minor** / **major**. The workflow tags the commit and runs [GoReleaser](https://goreleaser.com/). If there is no tag yet, the first run uses **v0.1.0**.

Local snapshot build (Docker):

```bash
make release-snapshot
```

## Security

Do not commit API keys. Use CI secrets for automation.

## License

MIT — see [LICENSE](LICENSE).
