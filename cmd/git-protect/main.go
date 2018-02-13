package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
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
	install := flag.Bool("install", false, "Install the pre-push hook and exit")
	force := flag.BoolP("force", "f", false, "Force the branch to be protected even if missing the pre-push hook")
	flag.Parse()

	dir, err := gitDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %s", err)
		return 1
	}

	// Check for the existence of the pre-push hook.
	st, err := os.Stat(filepath.Join(dir, "hooks/pre-push"))
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "fatal: hooks/pre-push: %s\n", err)
		return 1
	}

	// Install the git hook if it was requested and exit.
	if *install {
		if err != nil {
			if err := os.Mkdir(filepath.Join(dir, "hooks"), 0777); err != nil && !os.IsExist(err) {
				fmt.Fprintf(os.Stderr, "fatal: Could not create .git/hooks directory: %s\n", err)
				return 1
			}

			if err := ioutil.WriteFile(filepath.Join(dir, "hooks/pre-push"), []byte(prePushHook), 0775); err != nil {
				fmt.Fprintf(os.Stderr, "fatal: Could not create pre-push hook: %s\n", err)
				return 1
			}
		}
		return 0
	}

	// If the hook is not present and we are not using force, then abort.
	if err != nil && !*force {
		fmt.Fprintf(os.Stderr, "fatal: No pre-push hook found in the .git directory, aborting\n")
		return 1
	} else if st.Mode()&0500 != 0500 && !*force {
		// If the hook is not executable, then also abort since the pre-push hook may not run correctly.
		fmt.Fprintf(os.Stderr, "fatal: pre-push hook is not executable and will not be run, aborting\n")
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

const prePushHook = `#!/bin/bash

git_dir=$(git rev-parse --git-dir)
[ $? -ne 0 ] && exit 1

while read local_ref local_sha remote_ref remote_sha
do
  # Retrieve the basename from the remote ref and ensure it doesn't
  # match one of the protected branches.
  remote_branch=$(basename $remote_ref)
  if [ -e "$git_dir/protected-branches/$remote_branch" ]; then
    echo >&2 "Remote branch is protected, not pushing"
    exit 1
  fi
done

exit 0
`
