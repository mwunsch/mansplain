# CLAUDE.md

mansplain generates mdoc(7) man pages from --help output and READMEs using an LLM.

## Build and test

```
go build -o mansplain .
go test ./...
```

Binary is `mansplain`. Config lives at `~/.config/mansplain/config.toml`.

## How it works

The `generate` command gathers source material (--help output, README content, stdin),
constructs a prompt with the system prompt from `SKILL.md`, sends it to an
OpenAI-compatible chat completions API, strips any markdown code fences from the response,
and writes the mdoc source to stdout (or a file with `-o`).

Tool name inference: if `--name` is not given, the tool tries to extract it from the
source text by looking for `Usage: toolname` patterns (same-line and next-line variants).
If `--name` is given but no source material is provided, it runs `<name> --help` automatically.

## Architecture

```
main.go                    Entrypoint, embeds SKILL.md, extracts system prompt from markers
cmd/
  root.go                  Global flags: --base-url, --api-key, --model
  generate.go              Main command. Sources: positional file, -, --from-help, --name
  configure.go             Interactive config setup via huh form
  install.go               Install man page to ~/.local/share/man/
  lint.go                  mdoc validation via mandoc -Tlint + section checks
internal/
  config/config.go         Load/save config, resolve flags > env > file > defaults
  llm/client.go            OpenAI chat completions client, returns content + Usage stats
  llm/prompt.go            System prompt storage, user prompt construction from sources
  llm/client_test.go       Mock HTTP server, golden file validation, code fence stripping
  ui/ui.go                 Styled stderr output via lipgloss (Catppuccin Mocha palette)
  ui/spinner.go            bubbletea spinner with elapsed time (falls back to static on non-TTY)
  ui/configure.go          huh form for mansplain configure
SKILL.md                   Agent skill AND system prompt (single source of truth)
testdata/help/grep.txt     Sample --help for tests
testdata/golden/grep.1     Expected mdoc output for golden file test
```

## Config resolution

Priority (highest first):
1. CLI flags (`--base-url`, `--api-key`, `--model`)
2. `MANSPLAIN_BASE_URL`, `MANSPLAIN_API_KEY`, `MANSPLAIN_MODEL` env vars
3. `~/.config/mansplain/config.toml` (base_url, api_key, model fields)
4. `OPENAI_API_KEY` env var (fallback for api_key only)
5. Default base URL: `https://api.openai.com/v1`

## Why mdoc(7)

Output is always mdoc semantic macros (`.Nm`, `.Fl`, `.Ar`, `.Sh`, `.Op`, `.Bl`/`.It`/`.El`),
never raw troff (`.ft`, `.sp`, `.in`, `.br`). mdoc is what OpenBSD's mandoc uses. It's
portable across mandoc, groff, and nroff, parseable by tools, and self-documenting at
the source level.

## The system prompt

`SKILL.md` is both the agent skill definition AND the system prompt source.
The CLI embeds SKILL.md and extracts the content between `<!-- system-prompt:start -->`
and `<!-- system-prompt:end -->` markers for the LLM system prompt. This ensures
the mdoc guidance is always in sync whether you're using the CLI or the agent skill.

The system prompt section uses a complete grep(1) man page as a few-shot example.
Small models (3-7B) learn the format much better from a concrete example than from
a reference card. Changes to SKILL.md directly affect both the agent skill and CLI
output quality. Test changes against both small local models and larger API models.

## UI conventions

- Man page content goes to stdout. Everything else goes to stderr.
- All stderr output uses lipgloss for styling (Catppuccin Mocha colors).
- The spinner uses bubbletea directly (not huh/spinner) to show elapsed time.
- On non-TTY stderr, the spinner falls back to a single static line.
- Token usage (prompt/completion/total + elapsed time) prints after generation.
- Errors use `ui.Error()` which renders a styled `✗` prefix.

## Code conventions

- `gofmt`, `go vet`
- Wrap errors with context: `fmt.Errorf("reading README: %w", err)`
- Cobra `RunE` for commands that can fail
- `SilenceUsage: true`, `SilenceErrors: true` on root; main.go handles error display
- Golden file tests with mock HTTP server for LLM client
- No global mutable state except the system prompt (set once at startup)

