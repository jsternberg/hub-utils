package main

import (
	"fmt"
	"os"
	"strings"

	"bufio"
	"context"
	"os/exec"
	"regexp"

	flag "github.com/spf13/pflag"
)

var reReflogCommitMessage = regexp.MustCompile(`^checkout: moving from ([^\s]+) to ([^\s]+)$`)

type CheckoutEvent struct {
	From, To string
}

func listBranches() (map[string]struct{}, error) {
	cmd := exec.Command("git", "branch", "--list")
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	branches := make(map[string]struct{})
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		branch := scanner.Text()
		if len(branch) > 0 && branch[0] == '*' {
			branch = branch[1:]
		}
		branch = strings.TrimSpace(branch)
		branches[branch] = struct{}{}
	}

	out.Close()
	if err := cmd.Wait(); err != nil {
		return nil, err
	}
	return branches, nil
}

func reflogMessages(ctx context.Context) (<-chan CheckoutEvent, error) {
	cmd := exec.Command("git", "log", "-g", "--grep-reflog=checkout: moving from", "--pretty=format:%gs")
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	ch := make(chan CheckoutEvent, 100)
	go func() {
		// We do not care about the exit status.
		defer cmd.Wait()
		defer out.Close()
		defer close(ch)

		// Scan the output from the log.
		scanner := bufio.NewScanner(out)
		for scanner.Scan() {
			m := reReflogCommitMessage.FindStringSubmatch(scanner.Text())
			if m == nil {
				continue
			}

			select {
			case ch <- CheckoutEvent{
				From: m[1],
				To:   m[2],
			}:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch, nil
}

func realMain() int {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: git-history

List the previous branches that have been visited within this workspace.
`)
	}
	flag.Parse()

	valid, err := listBranches()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not list branches: %s\n", err)
		return 1
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	seen := make(map[string]struct{})
	branches := make([]string, 0, 100)

	events, err := reflogMessages(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not read the reflog: %s\n", err)
		return 1
	}

	appendBranch := func(name string) {
		if _, ok := valid[name]; !ok {
			return
		}
		if _, ok := seen[name]; !ok {
			branches = append(branches, name)
			seen[name] = struct{}{}
		}
	}

	for e := range events {
		appendBranch(e.To)
		appendBranch(e.From)
	}

	for i := len(branches) - 1; i >= 0; i-- {
		fmt.Println(branches[i])
	}
	return 0
}

func main() {
	os.Exit(realMain())
}
