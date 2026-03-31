# Reusable Rule

Distribution and reuse of Migration Rules.

## Usage

You can run a Reusable Rule without a configuration file.
It's similar to running something with `npx`.
This is convenient for one-off executions.

```sh
yamledit run @<migration>
```

If you want to add it to a configuration file for version control or run it with an alias, you can add a Reusable Rule to the configuration file with `yamledit add`.

```sh
yamledit add [-g] [-f] [<alias>] [<migration>]
```

By default, it is added to the project configuration, but with the `-g` option it can be added to the global configuration.
If no alias is specified, the default name of the Reusable Rule is used.
If the specified alias already exists, an error is returned.
With the `-f` option, the existing alias is overwritten.
If `<migration>` is omitted, a Fuzzy Finder is launched and the selected rule is added.

Once added, you can run it by specifying the alias. Also, if added to the project configuration, it will be run by default with `yamledit run`.

```sh
yamledit run [@<alias>]
```

## Search

You can search for Reusable Rules with `yamledit search`.
Search results are output to stdout.

```sh
yamledit search [<query>]
```

If no arguments are specified, a Fuzzy Finder is launched and the selected Reusable Rule is output.

```
URL License description
```

If a query is specified, a list of fuzzy search matches is output.
AI Agents may find the query argument useful since interactive UIs like Fuzzy Finder are difficult for them to handle.

## Distribution

Distribution and publishing is straightforward.
Simply publish the rule at any location and make it downloadable via URL, or make it available through the GitHub Contents API.

## Discoverability

Simply publishing a rule may not make it easy for users to find.
Adding the topic `yamledit-rule` to the GitHub Repository that publishes the Reusable Rule makes it searchable via `yamledit search`.

https://github.com/search?q=topic%3Ayamledit-rule&type=repositories

```sh
gh api \
  -H "Accept: application/vnd.github+json" \
  -H "X-GitHub-Api-Version: 2026-03-10" \
  "/search/repositories?q=topic:yamledit-rule"
```

References:

- https://docs.github.com/en/rest/search/search?apiVersion=2026-03-10#search-repositories
- https://docs.github.com/en/search-github/searching-on-github/searching-for-repositories#search-by-topic

Since the description and license information displayed in search results are retrieved from the repository, it is more convenient to manage rules in a dedicated repository.
To make rules published outside of GitHub searchable, add the Reusable Rule to suzuki-shunsuke/yamledit-registry.

standard registry: https://github.com/suzuki-shunsuke/yamledit-registry

e.g.

```yaml
rules:
  - name: goreleaser-v2
    description: Upgrade .goreleaser.yaml to v2
    license: MIT
    license_url: https://example.com/goreleaser-v2/license
    url: https://example.com/goreleaser-v2/yamledit.yaml
  - name: checkout-disable-persist-credential
    description: Disable actions/checkout's persist-credentials
    github_content:
      owner: suzuki-shunsuke
      repo: yamledit
      ref: main
      path: examples/checkout-disable-persist-credential.yaml
```

yamledit caches the contents of the standard registry and topic search results for a certain period. See [cache.md](cache.md) for details.
