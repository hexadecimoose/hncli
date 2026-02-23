# hncli

A Hacker News CLI with an interactive TUI browser and plain-text output mode.

## Install

```sh
go install github.com/hexadecimoose/hncli/cmd/hncli@latest
```

Or build from source:

```sh
git clone https://github.com/hexadecimoose/hncli
cd hncli
make install
```

## Usage

```
hncli [command] [flags]
```

Run with no arguments to launch the interactive TUI browser (top stories).

### Commands

| Command | Description |
|---|---|
| `hncli` | Interactive TUI browser (top stories) |
| `hncli top` | Top stories |
| `hncli new` | Newest stories |
| `hncli best` | Best stories |
| `hncli ask` | Ask HN |
| `hncli show` | Show HN |
| `hncli jobs` | Job postings |
| `hncli item <id>` | Story and comments |
| `hncli user <name>` | User profile and recent submissions |
| `hncli search <query>` | Search via Algolia HN |

### Flags

| Flag | Description |
|---|---|
| `-n`, `--count` | Number of stories to fetch (default 30) |
| `-p`, `--plain` | Plain text output — no TUI. Auto-enabled when stdout is not a TTY |
| `--version` | Print version |

### TUI keybindings

**Story list**

| Key | Action |
|---|---|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `g` / `G` | Jump to top / bottom |
| `enter` | Open comments |
| `o` | Open story URL in browser |
| `c` | Open on news.ycombinator.com |
| `q` | Quit |

**Comments / User profile**

| Key | Action |
|---|---|
| `↑` / `k` | Scroll up |
| `↓` / `j` | Scroll down |
| `g` / `G` | Jump to top / bottom |
| `o` | Open story URL in browser |
| `c` | Open on news.ycombinator.com |
| `←` / `esc` / `backspace` | Back to list |
| `q` | Quit |

### Plain-text / scripting

`--plain` (or `-p`) prints to stdout instead of launching the TUI.
It is also auto-enabled when stdout is not a TTY, so piping just works:

```sh
hncli top -n 5 --plain
hncli search golang | grep -i generics
hncli item 12345678 --plain | less
```

## Data sources

- Stories and comments: [HN Firebase API](https://github.com/HackerNews/API)
- Search: [Algolia HN Search API](https://hn.algolia.com/api)

## Configuration

### `HNCLI_OPEN`

Controls what happens when you press `o` (open URL) or `c` (open HN discussion).
Use `{}` as a placeholder for the URL; if absent the URL is appended as the last argument.
The value is executed via `sh -c`, so pipes and redirects work.

```sh
# Default — system browser (xdg-open / open / rundll32)
unset HNCLI_OPEN

# Linux clipboard (X11)
export HNCLI_OPEN="echo {} | xclip -selection clipboard"

# Wayland clipboard
export HNCLI_OPEN="echo {} | wl-copy"

# macOS clipboard
export HNCLI_OPEN="echo {} | pbcopy"

# Explicit browser
export HNCLI_OPEN="firefox"
```

## License

[MIT](LICENSE)
