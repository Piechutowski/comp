package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: comp <cli-name> [shell]")
		os.Exit(1)
	}

	name := os.Args[1]

	if name == "completion" {
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: comp completion <bash|zsh|fish>")
			os.Exit(1)
		}
		script, err := completionScript(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Print(script)
		return
	}

	shell := ""
	if len(os.Args) >= 3 {
		shell = os.Args[2]
	} else {
		shell = detectShell()
	}
	if shell == "" {
		fmt.Fprintln(os.Stderr, "could not detect shell from $SHELL, pass it explicitly: comp <cli-name> <bash|zsh|fish>")
		os.Exit(1)
	}

	out, err := exec.Command(name, "completion", shell).Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to run %q completion %s: %v\n", name, shell, err)
		os.Exit(1)
	}

	path, err := targetPath(name, shell)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create directory %s: %v\n", filepath.Dir(path), err)
		os.Exit(1)
	}

	if err := os.WriteFile(path, out, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write completion file %s: %v\n", path, err)
		os.Exit(1)
	}

	fmt.Printf("Installed %s completions for %s -> %s\n", shell, name, path)

	switch shell {
	case "zsh":
		fmt.Println("Make sure ~/.zsh/completions is in your $fpath before compinit runs, e.g. in ~/.zshrc:")
		fmt.Println(`  fpath+=(~/.zsh/completions)`)
		fmt.Println(`  autoload -U compinit && compinit`)
	}

	fmt.Println("Restart your shell (or open a new tab) to pick up the changes.")
}


func completionScript(shell string) (string, error) {
	switch shell {
	case "fish":
		return `complete -c comp -f
complete -c comp -n '__fish_is_first_arg' -a '(__fish_complete_command)' -d 'CLI name'
complete -c comp -n 'not __fish_is_first_arg' -a 'bash zsh fish' -d 'shell'
`, nil
	case "bash":
		return `_comp_completions() {
    local cur
    cur="${COMP_WORDS[COMP_CWORD]}"
    if [ "$COMP_CWORD" -eq 1 ]; then
        COMPREPLY=( $(compgen -c -- "$cur") )
    elif [ "$COMP_CWORD" -eq 2 ]; then
        COMPREPLY=( $(compgen -W "bash zsh fish" -- "$cur") )
    fi
}
complete -F _comp_completions comp
`, nil
	case "zsh":
		return `#compdef comp
_comp() {
    if (( CURRENT == 2 )); then
        _command_names -e
    elif (( CURRENT == 3 )); then
        local -a shells
        shells=(bash zsh fish)
        _describe 'shell' shells
    fi
}
_comp "$@"
`, nil
	default:
		return "", fmt.Errorf("unsupported shell %q (supported: bash, zsh, fish)", shell)
	}
}

func detectShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return ""
	}
	return filepath.Base(shell)
}

func targetPath(name, shell string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}

	switch shell {
	case "fish":
		configHome := os.Getenv("XDG_CONFIG_HOME")
		if configHome == "" {
			configHome = filepath.Join(home, ".config")
		}
		return filepath.Join(configHome, "fish", "completions", name+".fish"), nil
	case "bash":
		dataHome := os.Getenv("XDG_DATA_HOME")
		if dataHome == "" {
			dataHome = filepath.Join(home, ".local", "share")
		}
		return filepath.Join(dataHome, "bash-completion", "completions", name), nil
	case "zsh":
		return filepath.Join(home, ".zsh", "completions", "_"+name), nil
	default:
		return "", fmt.Errorf("unsupported shell %q (supported: bash, zsh, fish)", shell)
	}
}
