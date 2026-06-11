package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorCyan   = "\033[36m"
	colorYellow = "\033[33m"
	colorGreen  = "\033[32m"
)

func main() {
	if len(os.Args) != 2 {
		printUsage()
		os.Exit(1)
	}

	name := os.Args[1]

	out, err := exec.Command(name, "completion", "fish").Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to run %q completion fish: %v\n", name, err)
		os.Exit(1)
	}

	path, err := targetPath(name)
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

	fmt.Printf("%sInstalled%s fish completions for %s%s%s -> %s\n",
		colorGreen, colorReset, colorBold, name, colorReset, path)
	fmt.Println("Restart your shell (or open a new tab) to pick up the changes.")
}

func printUsage() {
	fmt.Printf("%s%susage:%s comp <cli-name>\n\n", colorBold, colorYellow, colorReset)
	fmt.Printf("Installs fish completions for %s<cli-name>%s, which must support a\n", colorCyan, colorReset)
	fmt.Printf("%s<cli-name> completion fish%s subcommand.\n", colorCyan, colorReset)
}

func targetPath(name string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}

	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, "fish", "completions", name+".fish"), nil
}
