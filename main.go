package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorCyan   = "\033[36m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	name := os.Args[1]

	var shell string
	if len(os.Args) >= 3 {
		shell = normalizeShell(os.Args[2])
		if shell == "" {
			fmt.Fprintf(os.Stderr, "%sunsupported shell %q (supported: bash, zsh, fish, powershell)%s\n", colorRed, os.Args[2], colorReset)
			os.Exit(1)
		}
	} else {
		shell = detectShell()
		if shell == "" {
			fmt.Fprintf(os.Stderr, "%scould not detect your shell, pass it explicitly: comp <cli-name> <bash|zsh|fish|powershell>%s\n", colorRed, colorReset)
			os.Exit(1)
		}
	}

	out, err := exec.Command(name, "completion", shell).Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sfailed to run %q completion %s: %v%s\n", colorRed, name, shell, err, colorReset)
		os.Exit(1)
	}

	path, err := targetPath(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s%v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "%sfailed to create directory %s: %v%s\n", colorRed, filepath.Dir(path), err, colorReset)
		os.Exit(1)
	}

	if err := os.WriteFile(path, out, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "%sfailed to write completion file %s: %v%s\n", colorRed, path, err, colorReset)
		os.Exit(1)
	}

	fmt.Printf("Installed %s completions for %s -> %s\n", shell, name, path)

	switch shell {
	case "zsh":
		fmt.Println("Make sure ~/.zsh/completions is in your $fpath before compinit runs, e.g. in ~/.zshrc:")
		fmt.Println(`  fpath+=(~/.zsh/completions)`)
		fmt.Println(`  autoload -U compinit && compinit`)
	case "powershell":
		fmt.Printf("Add this line to your PowerShell profile (%s):\n", "$PROFILE")
		fmt.Printf("  . %s\n", path)
	}

	fmt.Println("Restart your shell (or open a new tab) to pick up the changes.")
}

func printUsage() {
	fmt.Printf("%s%scomp%s - install shell completions for any CLI\n\n", colorBold, colorCyan, colorReset)
	fmt.Printf("%sUsage:%s\n", colorBold, colorReset)
	fmt.Printf("  %scomp <cli-name> [shell]%s\n\n", colorYellow, colorReset)
	fmt.Println("  <cli-name>  the CLI to install completions for (must support `<cli-name> completion <shell>`)")
	fmt.Println("  [shell]     bash, zsh, fish, or powershell (auto-detected if omitted)")
}

// detectShell figures out which shell to install completions for, so that
// `comp <cli>` just works without passing a shell explicitly.
//
// Detection order:
//  1. The parent process (the shell you actually typed the command in) — most
//     accurate, works on Linux, macOS and Windows.
//  2. The $SHELL environment variable (your login shell) — Unix only.
//  3. An OS-aware default: Windows → powershell.
func detectShell() string {
	if s := normalizeShell(parentProcessName()); s != "" {
		return s
	}

	if shell := os.Getenv("SHELL"); shell != "" {
		if s := normalizeShell(filepath.Base(shell)); s != "" {
			return s
		}
	}

	if runtime.GOOS == "windows" {
		return "powershell"
	}

	return ""
}

// parentProcessName returns the executable name of the process that launched
// comp (typically your interactive shell), or "" if it can't be determined.
func parentProcessName() string {
	ppid := os.Getppid()

	// Linux: read the parent's name from procfs.
	if runtime.GOOS == "linux" {
		if comm, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", ppid)); err == nil {
			return strings.TrimSpace(string(comm))
		}
	}

	// macOS / other Unix: procfs is unavailable, so ask ps for the parent's
	// command name.
	if runtime.GOOS != "windows" {
		if out, err := exec.Command("ps", "-o", "comm=", "-p", fmt.Sprintf("%d", ppid)).Output(); err == nil {
			return filepath.Base(strings.TrimSpace(string(out)))
		}
	}

	return ""
}

// normalizeShell maps shell executable names to the canonical shell name used
// by CLI completion subcommands.
func normalizeShell(name string) string {
	// Login shells are often reported with a leading dash (e.g. "-zsh").
	name = strings.TrimPrefix(name, "-")
	switch strings.ToLower(name) {
	case "bash":
		return "bash"
	case "zsh":
		return "zsh"
	case "fish":
		return "fish"
	case "powershell", "powershell.exe", "pwsh", "pwsh.exe":
		return "powershell"
	}
	return ""
}

func targetPath(name string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}

	switch shell {
	case "fish":
		return filepath.Join(xdgConfigHome(home), "fish", "completions", name+".fish"), nil
	case "bash":
		dataHome := os.Getenv("XDG_DATA_HOME")
		if dataHome == "" {
			dataHome = filepath.Join(home, ".local", "share")
		}
		return filepath.Join(dataHome, "bash-completion", "completions", name), nil
	case "zsh":
		return filepath.Join(home, ".zsh", "completions", "_"+name), nil
	case "powershell":
		return powershellCompletionsDir(home, name)
	default:
		return "", fmt.Errorf("unsupported shell %q (supported: bash, zsh, fish, powershell)", shell)
	}
}

// powershellCompletionsDir returns the per-platform path for PowerShell
// completion scripts. On Windows this is under Documents\PowerShell; on
// Linux/Mac (pwsh) it follows the XDG config convention.
func powershellCompletionsDir(home, name string) (string, error) {
	if runtime.GOOS == "windows" {
		return filepath.Join(home, "Documents", "PowerShell", "Completions", name+".ps1"), nil
	}
	return filepath.Join(xdgConfigHome(home), "powershell", "Completions", name+".ps1"), nil
}

// xdgConfigHome returns $XDG_CONFIG_HOME, defaulting to ~/.config.
func xdgConfigHome(home string) string {
	if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
		return configHome
	}
	return filepath.Join(home, ".config")
}
