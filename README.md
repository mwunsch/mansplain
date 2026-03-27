# mansplain

Generate man pages from `--help` output and READMEs using an LLM.

```
mansplain generate --name rg -o rg.1
man ./rg.1
```

mansplain takes source material you already have and produces well-formed
[mdoc(7)](https://man.openbsd.org/mdoc.7) man pages. It targets any
OpenAI-compatible API, so it works with OpenAI, local models via
LM Studio or Ollama, or any other compatible endpoint.

## Install

```
go install github.com/mwunsch/mansplain@latest
```

Or build from source:

```
git clone https://github.com/mwunsch/mansplain.git
cd mansplain
go build -o mansplain .
```

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

# From --help output
mansplain generate --from-help "rg --help" -o rg.1

# From a README
mansplain generate --from-readme README.md --name mytool

# From stdin
curl --help | mansplain generate - --name curl

# Multiple sources for richer output
mansplain generate --from-help "rg --help" --from-readme README.md -o rg.1
```

Preview with mandoc:

```
mansplain generate --name jq | mandoc -Tutf8 | less
```

## Commands

| Command | Description |
|---------|-------------|
| `generate` | Generate a man page from source material via LLM |
| `configure` | Interactively set up the LLM connection |
| `scaffold` | Generate an mdoc template offline (no LLM) |
| `install` | Install a man page to the system MANPATH |
| `preview` | Render a man page in the terminal |
| `lint` | Validate man page structure |

## Configuration

mansplain reads configuration from (highest priority first):

1. CLI flags (`--base-url`, `--api-key`, `--model`)
2. Environment variables (`MANSPLAIN_BASE_URL`, `MANSPLAIN_API_KEY`, `MANSPLAIN_MODEL`)
3. Config file (`~/.config/mansplain/config.toml`)
4. `OPENAI_API_KEY` environment variable
5. Default: `https://api.openai.com/v1`

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
