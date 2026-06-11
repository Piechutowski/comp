
# comp

A tiny tool that installs fish completions for any CLI that supports a
`<name> completion fish` subcommand, the convention used by
[urfave/cli](https://github.com/urfave/cli) and several other CLI frameworks.

## Usage

```sh
comp cli
```

This will:

1. Run `cli completion fish` to generate the completion script.
2. Write it to `$XDG_CONFIG_HOME/fish/completions/<name>.fish`,
   defaulting to `~/.config/fish/completions/<name>.fish`.

Restart your shell (or open a new tab) afterwards to pick up the changes.
