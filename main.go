package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorCyan   = "\033[36m"
	colorYellow = "\033[33m"
	colorGreen  = "\033[32m"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	name := os.Args[1]

	shell := ""
	if len(os.Args) >= 3 {
		shell = os.Args[2]
	} else {
		shell = detectShell()
	}
	if shell == "" {
		fmt.Fprintln(os.Stderr, "could not detect shell, pass it explicitly: comp <cli-name> <bash|zsh|fish>")
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

	fmt.Printf("%sInstalled%s %s%s%s completions for %s%s%s -> %s\n",
		colorGreen, colorReset, colorBold, shell, colorReset, colorBold, name, colorReset, path)

	switch shell {
	case "zsh":
		fmt.Println("Make sure ~/.zsh/completions is in your $fpath before compinit runs, e.g. in ~/.zshrc:")
		fmt.Println(`  fpath+=(~/.zsh/completions)`)
		fmt.Println(`  autoload -U compinit && compinit`)
	}

	fmt.Println("Restart your shell (or open a new tab) to pick up the changes.")
}

func printUsage() {
	fmt.Printf("%s%susage:%s comp <cli-name> [shell]\n\n", colorBold, colorYellow, colorReset)
	fmt.Printf("Installs shell completions for %s<cli-name>%s, which must support a\n", colorCyan, colorReset)
	fmt.Printf("%s<cli-name> completion <shell>%s subcommand.\n\n", colorCyan, colorReset)
	fmt.Printf("%sshell%s defaults to the shell you're currently running, detected\n", colorCyan, colorReset)
	fmt.Println("automatically (override with bash, zsh, or fish).")
}

// detectShell tries to determine the current shell, first from $SHELL and
// falling back to inspecting the parent process (so it works even when
// $SHELL doesn't reflect the shell actually invoking comp).
func detectShell() string {
	if shell := os.Getenv("SHELL"); shell != "" {
		return filepath.Base(shell)
	}

	if exe, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", os.Getppid())); err == nil {
		return filepath.Base(exe)
	}

	if data, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", os.Getppid())); err == nil {
		return strings.TrimSpace(string(data))
	}

	return ""
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
