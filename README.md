# Hub Utils

This repository contains utility tools for GitHub. It is based around [hub](https://github.com/github/hub) and is meant to be used as an extension of that command, although it can be used separately.

Since this tool uses hub as a library, it integrates well with the tool itself by keeping credentials in the same area as the hub tool.

## Tools

### git rate-limit

This will output the rate limit information for the GitHub API using the account information configured with hub. See the [documentation](https://developer.github.com/v3/#rate-limiting) for details on rate limiting in the GitHub API.

### git conflicts

This will output the pull requests for a repository that have merge conflicts for the current user. It will only output pull requests from the current user and not all conflicts. This may occasionally need to be run twice because the GitHub API doesn't always return the correct thing every time with merge conflicts.
