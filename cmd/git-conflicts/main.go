package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/github/hub/github"
	"github.com/github/hub/utils"
)

func main() {
	localRepo, err := github.LocalRepo()
	utils.Check(err)

	project, err := localRepo.MainProject()
	utils.Check(err)

	config := github.CurrentConfig()
	host, err := config.PromptForHost(project.Host)
	if err != nil {
		utils.Check(github.FormatError("checking pull request conflicts", err))
	}

	params := url.Values{}
	params.Set("state", "open")
	params.Set("per_page", "100")
	u := url.URL{
		Scheme:   host.Protocol,
		Host:     host.Host,
		Path:     fmt.Sprintf("/repos/%s/%s/pulls", project.Owner, project.Name),
		RawQuery: params.Encode(),
	}
	if u.Host == github.GitHubHost {
		u.Host = github.GitHubApiHost
	}
	req, _ := http.NewRequest("GET", u.String(), nil)
	req.Header.Set("User-Agent", github.UserAgent)
	req.Header.Set("Authorization", fmt.Sprintf("token %s", host.AccessToken))

	resp, err := http.DefaultClient.Do(req)
	utils.Check(err)

	type pullRequestInfo struct {
		Url  string `json:"url"`
		User struct {
			Login string `json:"login"`
		} `json:"user"`
	}
	var pullRequests []pullRequestInfo
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&pullRequests); err != nil {
		utils.Check(github.FormatError("decoding pull request information", err))
	}

	for _, pr := range pullRequests {
		if pr.User.Login != host.User {
			continue
		}

		req, _ := http.NewRequest("GET", pr.Url, nil)
		req.Header.Set("User-Agent", github.UserAgent)
		req.Header.Set("Authorization", fmt.Sprintf("token %s", host.AccessToken))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			utils.Check(github.FormatError("fetching pull request", err))
		}

		var pullRequest struct {
			Head struct {
				Label string `json:"label"`
				Ref   string `json:"ref"`
				User  struct {
					Login string `json:"login"`
				} `json:"user"`
			} `json:"head"`
			Base struct {
				Label string `json:"label"`
				Ref   string `json:"ref"`
				User  struct {
					Login string `json:"login"`
				} `json:"user"`
			} `json:"base"`
			Mergeable *bool `json:"mergeable"`
		}
		dec := json.NewDecoder(resp.Body)
		if err := dec.Decode(&pullRequest); err != nil {
			utils.Check(github.FormatError("decoding pull request details", err))
		}

		if pullRequest.Mergeable == nil || *pullRequest.Mergeable {
			continue
		}

		from := pullRequest.Head.Label
		if pullRequest.Head.User.Login == project.Owner {
			from = pullRequest.Head.Ref
		}
		to := pullRequest.Base.Label
		if pullRequest.Base.User.Login == project.Owner {
			to = pullRequest.Base.Ref
		}
		fmt.Printf("%s to %s\n", from, to)
	}
}
