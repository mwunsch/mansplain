# mansplain

Generate man pages from `--help` output and READMEs using an LLM.

```
mansplain generate --name rg -o rg.1
man ./rg.1
```

Most CLI tools ship without man pages. The format is old, the toolchain
is hostile, and writing mdoc by hand is nobody's idea of a good time.
mansplain fixes that. Point it at a `--help` output or a README and it
produces a well-formed [mdoc(7)](https://man.openbsd.org/mdoc.7) man page.

It also ships as an agent skill, so any AI coding agent can generate
man pages as part of its normal workflow. The goal is to make man pages
a standard part of the CLI project scaffold, like README.md already is.

mansplain targets any OpenAI-compatible API: OpenAI, local models via
LM Studio or Ollama, or any other compatible endpoint.

## Install

```
curl -fsSL https://raw.githubusercontent.com/mwunsch/mansplain/main/install.sh | sh
```

This downloads the latest release, installs the binary to `~/.local/bin`
and the man page to `~/.local/share/man/man1`. Override with `INSTALL_DIR`
and `MAN_DIR` environment variables.

Or with Go (binary only, no man page):

```
go install github.com/mwunsch/mansplain@latest
```

Or download a tarball directly from
[GitHub Releases](https://github.com/mwunsch/mansplain/releases).

## Quick start

Configure your LLM connection:

```
mansplain configure
```

This prompts for a base URL, API key, and model, then saves to
`~/.config/mansplain/config.toml`. Defaults to OpenAI's API.

For local models, point it at LM Studio or Ollama:

```toml
# ~/.config/mansplain/config.toml
base_url = "http://localhost:1234/v1"
api_key = "lm-studio"
```

Generate a man page:

```
# From a tool name (runs --help automatically)
mansplain generate --name jq -o jq.1

# From a README
mansplain generate README.md --name mytool

# From --help output
mansplain generate --from-help "rg --help" -o rg.1

# From stdin
curl --help | mansplain generate - --name curl
```

Preview and validate:

```
mansplain generate --name jq | mandoc -Tutf8 | less
mansplain generate --name jq | mansplain lint -
```

## Agent skill

This is the highest-leverage feature. mansplain ships as an
[Agent Skill](https://agentskills.io) that teaches any compatible coding
agent how to write proper mdoc(7) man pages. It works across Claude Code,
Cursor, Copilot, Gemini CLI, and
[30+ others](https://agentskills.io).

The skill works without the mansplain binary. The agent uses its own model
and the full project context to generate the man page directly. This
typically produces better results than a standalone LLM call because the
agent has access to the source code, not just the help text.

Install the skill:

```
npx skills add mwunsch/mansplain
```

Or copy `SKILL.md` from this repo into your project's skills directory.

The skill teaches the agent to:
1. Read the project's README and CLI help output
2. Write a complete mdoc(7) man page following mdoc conventions
3. Validate with `mandoc -Tlint` (or `mansplain lint` if installed)
4. Place the file at `man/<toolname>.1`

The goal: every CLI project built with an AI coding agent should get a
man page as part of the standard scaffold. If agents know how to write
README.md, they should know how to write a man page too.

## Commands

| Command | Description |
|---------|-------------|
| `generate` | Generate a man page from source material via LLM |
| `lint` | Validate man page structure and completeness |
| `install` | Install a man page so man(1) can find it |
| `configure` | Interactively set up the LLM connection |

Use `generate --dry-run` to see the assembled prompt without calling the API.

## Configuration

mansplain reads configuration from (highest priority first):

1. CLI flags (`--base-url`, `--api-key`, `--model`)
2. Environment variables (`MANSPLAIN_BASE_URL`, `MANSPLAIN_API_KEY`, `MANSPLAIN_MODEL`)
3. Config file (`~/.config/mansplain/config.toml`)
4. `OPENAI_API_KEY` environment variable
5. Default: `https://api.openai.com/v1`

## Working with man pages

Once you have a man page, here's how to use the existing toolchain:

```
# Preview a local man page file
mandoc -Tutf8 rg.1 | less

# Install it so man(1) can find it
mansplain install rg.1

# Now it works like any other man page
man rg

# Search installed man pages by keyword
apropos "regular expression"

# One-line description of a tool
whatis grep

# See where man looks for pages
manpath
```

`mansplain install` copies man pages to `~/.local/share/man/`, which is
in the default search path on both macOS and Linux. No sudo required.

If you need to add a custom directory to the man search path, set `MANPATH`:

```
export MANPATH="$HOME/myproject/man:$(manpath)"
```

Note: `man -l file.1` works on Linux for previewing local files but not
on macOS. Use `mandoc -Tutf8 file.1 | less` instead, which works everywhere.

## Output format

mansplain produces [mdoc(7)](https://man.openbsd.org/mdoc.7) source, the
semantic macro set used by OpenBSD's mandoc and widely supported by groff.
mdoc uses semantic markup (`.Nm` for name, `.Fl` for flag, `.Ar` for argument)
rather than raw formatting directives, making the output portable and
machine-parseable.

## Model quality

Output quality depends on the model. Larger models (GPT-4o, Claude) produce
clean, valid mdoc with correct section structure and flag documentation.
Smaller local models (3-7B parameters) get the general structure right but
may have syntax errors or hallucinate flags. The system prompt uses a
few-shot example to maximize compatibility with smaller models.

## License

MIT
