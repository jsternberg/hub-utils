# Hub Utils

This repository contains utility tools for GitHub. It is based around [hub](https://github.com/github/hub) and is meant to be used as an extension of that command, although it can be used separately.

Since this tool uses hub as a library, it integrates well with the tool itself by keeping credentials in the same area as the hub tool.

## Tools

### git rate-limit

This will output the rate limit information for the GitHub API using the account information configured with hub. See the [documentation](https://developer.github.com/v3/#rate-limiting) for details on rate limiting in the GitHub API.

### git conflicts

This will output the pull requests for a repository that have merge conflicts for the current user. It will only output pull requests from the current user and not all conflicts. This may occasionally need to be run twice because the GitHub API doesn't always return the correct thing every time with merge conflicts.

### git history

This will read through the reflog to find the recent branches that have been visited. It automatically filters out any branches that have been deleted.

### git protect

This will mark a branch as protected on the client side. It does not mark a branch as protected on GitHub.

A pre-push script can then be used to check if a branch is protected to prevent an accidental client-side push.

### git prune-branch

This will delete branches where the remote is no longer present. This is most useful with pull requests and the "Delete Branch" button on GitHub. After merging a pull request and deleting the pull request using the button, you need to remember to also delete the local copy otherwise you can end up with too many branches that clutter the workspace and make it difficult to find active branches. This scans which branches have been merged into master, checks if the upstream branch still exists, and deletes the branch if it does not.

It will not delete any branches on the remote repository.
