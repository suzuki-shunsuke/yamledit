# GitHub Access Token

You can provide a GitHub Access Token to fetch migration files via the GitHub Contents API.
Public repositories can be accessed without a token, but you are more likely to hit API rate limits.
Private repositories require `contents:read` permission.
If you are not fetching migration files from GitHub, no token is needed.
There are several ways to provide an access token:

1. Environment variables
    1. `YAMLEDIT_GITHUB_TOKEN`
    1. `GITHUB_TOKEN`
1. [ghtkn integration](https://github.com/suzuki-shunsuke/ghtkn)
    1. Set the environment variable `YAMLEDIT_GHTKN_ENABLED` to `true`
    1. The ghtkn CLI is not required. However, a ghtkn configuration file is needed. Please refer to the ghtkn documentation for details.
