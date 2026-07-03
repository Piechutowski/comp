# comp

Install shell completions for any CLI that supports a
`<name> completion <shell>` subcommand (the convention used by
[urfave/cli](https://cli.urfave.org) and others).

Just run:

```sh
comp <cli-name>
```

`comp` detects your shell automatically and writes the completion script to the
right place. Works on Linux, macOS and Windows with bash, zsh, fish and
PowerShell. Pass a shell explicitly if you want to override detection:

```sh
comp <cli-name> [bash|zsh|fish|powershell]
```

## Install

```fish
go install github.com/piechutowski/comp@latest
```

Or, if you use [gobin](https://github.com/piechutowski/gobin):

```fish
gobin addr github.com/piechutowski/comp
```

If `~/go/bin` is not on your PATH, add it with `fish_add_path ~/go/bin`.
