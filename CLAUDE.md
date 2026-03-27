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
constructs a prompt with the system prompt from `prompts/generate.md`, sends it to an
OpenAI-compatible chat completions API, strips any markdown code fences from the response,
and writes the mdoc source to stdout (or a file with `-o`).

Tool name inference: if `--name` is not given, the tool tries to extract it from the
source text by looking for `Usage: toolname` patterns (same-line and next-line variants).
If `--name` is given but no source material is provided, it runs `<name> --help` automatically.

## Architecture

```
main.go                    Entrypoint, embeds prompts/generate.md via go:embed
cmd/
  root.go                  Global flags: --base-url, --api-key, --model
  generate.go              Main command. Sources: positional file, -, --from-help, --name
  configure.go             Interactive config setup via huh form
  scaffold.go              Offline mdoc template (not yet implemented)
  install.go               Man page installation (not yet implemented)
  preview.go               Terminal rendering via mandoc/groff (not yet implemented)
  lint.go                  mdoc validation via mandoc -Tlint (not yet implemented)
internal/
  config/config.go         Load/save config, resolve flags > env > file > defaults
  llm/client.go            OpenAI chat completions client, returns content + Usage stats
  llm/prompt.go            System prompt storage, user prompt construction from sources
  llm/client_test.go       Mock HTTP server, golden file validation, code fence stripping
  ui/ui.go                 Styled stderr output via lipgloss (Catppuccin Mocha palette)
  ui/spinner.go            bubbletea spinner with elapsed time (falls back to static on non-TTY)
  ui/configure.go          huh form for mansplain configure
prompts/generate.md        System prompt: few-shot mdoc example + concise rules
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

`prompts/generate.md` is the most important file in the project. It uses a complete
grep(1) man page as a few-shot example rather than describing mdoc rules abstractly.
Small models (3-7B) learn the format much better from a concrete example than from
a reference card. The prompt ends with a concise rules list for constraints the example
doesn't fully demonstrate.

Changes to this file directly affect output quality across all models. Test changes
against both small local models and larger API models.

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

## Stubs

`scaffold`, `install`, `preview`, and `lint` are registered as cobra commands but
return "not yet implemented" errors. They are described in README.md. When implementing:

- **scaffold**: generate a mdoc template with TODO placeholders, no LLM needed.
  Use `internal/mdoc/template.go`.
- **preview**: shell out to `mandoc -Tutf8 <file> | less`, fall back to
  `groff -mandoc -Tutf8`, last resort raw cat.
- **lint**: shell out to `mandoc -Tlint` if available, plus check for required
  sections (NAME, SYNOPSIS, DESCRIPTION) and flag consistency.
- **install**: detect MANPATH, copy file, run `mandb`/`makewhatis` if available.
  See `internal/manpath/`.
