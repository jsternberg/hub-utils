package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	flag "github.com/spf13/pflag"
)

func getDefaultBranch() string {
	cmd := exec.Command("git", "config", "git-prune.default-branch")
	out, err := cmd.Output()
	if err == nil {
		return string(bytes.TrimSpace(out))
	}
	return "master"
}

func hasRemote(branch string) bool {
	cmd := exec.Command("git", "config", fmt.Sprintf("branch.%s.remote", branch))
	return cmd.Run() == nil
}

func hasUpstream(branch string) bool {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", fmt.Sprintf("%s@{u}", branch))
	return cmd.Run() == nil
}

func deleteBranch(branch string) error {
	cmd := exec.Command("git", "branch", "-D", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func realMain() int {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: git-prune-branch <branch-name>

Delete the branches that have been merged into the target branch.
`)
	}
	flag.Parse()

	defaultBranch := getDefaultBranch()
	cmd := exec.Command("git", "branch", "--merged", defaultBranch)
	out, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not create pipe: %s\n", err)
		return 1
	}

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not list branches: %s\n", err)
		return 1
	}

	var exitcode int
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		branch := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(branch, "*") {
			// Skip a branch if we are currently on that branch.
			// We cannot remove this branch anyway.
			continue
		} else if branch == defaultBranch {
			// Skip past the default branch.
			continue
		}

		// Only remove remote branchesif they have a remote set so we only delete
		// branches that had a pull request created and merged for them.
		if hasRemote(branch) && !hasUpstream(branch) {
			if err := deleteBranch(branch); err != nil {
				exitcode++
			}
		}
	}
	out.Close()

	if err := cmd.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}
	return exitcode
}

func main() {
	os.Exit(realMain())
}
