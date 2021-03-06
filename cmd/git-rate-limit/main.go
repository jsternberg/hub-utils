package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/github/hub/github"
	"github.com/github/hub/utils"
	flag "github.com/spf13/pflag"
)

const GitHubApiHost = "api.github.com"

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: git-rate-limit

List the current rate limit information for the current user account.
`)
	}
	flag.Parse()

	localRepo, err := github.LocalRepo()
	utils.Check(err)

	project, err := localRepo.MainProject()
	utils.Check(err)

	config := github.CurrentConfig()
	host, err := config.PromptForHost(project.Host)
	if err != nil {
		utils.Check(github.FormatError("checking pull request conflicts", err))
	}

	u := url.URL{
		Scheme: host.Protocol,
		Host:   host.Host,
		Path:   fmt.Sprintf("/rate_limit", project.Owner, project.Name),
	}
	if u.Host == github.GitHubHost {
		u.Host = GitHubApiHost
	}
	req, _ := http.NewRequest("GET", u.String(), nil)
	req.Header.Set("User-Agent", github.UserAgent)
	req.Header.Set("Authorization", fmt.Sprintf("token %s", host.AccessToken))

	resp, err := http.DefaultClient.Do(req)
	utils.Check(err)

	resetTime, _ := strconv.ParseInt(resp.Header.Get("X-RateLimit-Reset"), 10, 64)
	fmt.Printf("Total:      %s\n", resp.Header.Get("X-RateLimit-Limit"))
	fmt.Printf("Remaining:  %s\n", resp.Header.Get("X-RateLimit-Remaining"))
	fmt.Printf("Reset Time: %s\n", time.Unix(resetTime, 0))
}
