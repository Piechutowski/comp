
# comp

A tiny tool that installs shell completions for any CLI that supports a
`<name> completion <shell>` subcommand (the convention used by
[urfave/cli](https://github.com/urfave/cli) and several other CLI frameworks).

Instead of manually figuring out where your shell expects completion files
and redirecting output by hand, just run:

```sh
comp <cli-name>
```

## Usage

```sh
comp cli
```

This will:

1. Detect your current shell automatically (no flags needed).
2. Run `cli completion <shell>` to generate the completion script.
3. Write it to the appropriate location for your shell:
   - **fish**: `$XDG_CONFIG_HOME/fish/completions/<name>.fish`
     (defaults to `~/.config/fish/completions/<name>.fish`)
   - **bash**: `$XDG_DATA_HOME/bash-completion/completions/<name>`
     (defaults to `~/.local/share/bash-completion/completions/<name>`)
   - **zsh**: `~/.zsh/completions/_<name>`

Restart your shell (or open a new tab) afterwards to pick up the changes.

### Specifying the shell explicitly

```sh
comp cli fish
```

### zsh setup

For zsh, make sure `~/.zsh/completions` is in your `$fpath` before
`compinit` runs. Add this to your `~/.zshrc`:

```sh
fpath+=(~/.zsh/completions)
autoload -U compinit && compinit
```

## Building

```sh
go build -o comp .
```

## Requirements

The target CLI must support a `completion <shell>` subcommand that prints
a completion script to stdout (`bash`, `zsh`, and `fish` are supported).
