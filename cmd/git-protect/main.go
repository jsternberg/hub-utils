package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	flag "github.com/spf13/pflag"
)

func gitDir() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.New(strings.TrimPrefix("fatal: ", strings.TrimSpace(string(out))))
	}
	return string(bytes.TrimSpace(out)), nil
}

func currentBranch() (string, error) {
	cmd := exec.Command("git", "symbolic-ref", "--short", "HEAD")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.New(strings.TrimPrefix("fatal: ", strings.TrimSpace(string(out))))
	}
	return string(bytes.TrimSpace(out)), nil
}

func realMain() int {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: git-protect [-u] <branch-name>

Protect a branch from being pushed to even if the push would normally be valid.

Options:
`)
		flag.PrintDefaults()
	}
	unprotect := flag.BoolP("unprotect", "u", false, "Remove the branch protection")
	flag.Parse()

	dir, err := gitDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %s", err)
		return 1
	}

	var branch string
	if flag.NArg() > 0 {
		branch = flag.Arg(0)
	} else {
		b, err := currentBranch()
		if err != nil {
			fmt.Fprintf(os.Stderr, "fatal: %s\n", err)
			return 1
		}
		branch = b
	}

	protectedBranchesDir := filepath.Join(dir, "protected-branches")
	if st, err := os.Stat(protectedBranchesDir); err == nil {
		if st.Mode()&os.ModeDir == 0 {
			fmt.Fprintf(os.Stderr, "fatal: %s is not a directory, aborting\n", protectedBranchesDir)
			return 1
		}
	} else if os.IsNotExist(err) {
		if err := os.Mkdir(protectedBranchesDir, 0777); err != nil {
			fmt.Fprintf(os.Stderr, "fatal: Could not make protected-branches directory: %s\n", err)
			return 1
		}
	} else {
		fmt.Fprintf(os.Stderr, "fatal: Could not access protected-branches directory: %s\n", err)
	}

	if !*unprotect {
		f, err := os.Create(filepath.Join(protectedBranchesDir, branch))
		if err != nil {
			fmt.Fprintf(os.Stderr, "fatal: Could not protect branch: %s\n", err)
			return 1
		}
		f.Close()
	} else {
		if err := os.Remove(filepath.Join(protectedBranchesDir, branch)); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "fatal: Could not unprotect branch: %s\n", err)
			return 1
		}
	}
	return 0
}

func main() {
	os.Exit(realMain())
}
