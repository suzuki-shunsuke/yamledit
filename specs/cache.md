# Cache

Migration files fetched from remote sources via the GitHub Contents API or URLs are cached for a certain period.

## Storage location

The cache storage location is determined in the following order of priority:

1. `${YAMLEDIT_CACHE_HOME}`
1. `${XDG_CACHE_HOME}/yamledit`
1. (Linux, macOS): `${HOME}/.cache/yamledit`
1. (Windows): `LocalAppData\cache\yamledit`
1. (Windows): `%LOCALAPPDATA%\cache\yamledit`

In the following sections, the cache directory is referred to as `${YAMLEDIT_CACHE_HOME}`.

The cache can be safely deleted as needed.

```sh
rm -Rf "${YAMLEDIT_CACHE_HOME}"
```

For URLs, the URL is base64 encoded.

```
${YAMLEDIT_CACHE_HOME}/remotes/url/<URL|base64 encode>/
  migration.yaml
  metadata.json
```

For the GitHub Contents API, the ref and path are base64 encoded.

```
${YAMLEDIT_CACHE_HOME}/remotes/github_content/github.com/<owner>/<repo>/<ref|base64 encode>/<path|base64 encode>/
  migration.yaml
  metadata.json
```

- `migration.yaml`: The downloaded migration file
- `metadata.json`: Metadata. Records the last updated time used for cache updates

metadata.json

```json
{
  "last_updated": "2006-01-02T15:04:05Z07:00"
}
```

- `last_updated`: RFC3339

## How it works

When `.rules[].import` is used, the cache is checked first.
If the cache exists, the last updated time recorded in metadata.json is checked to determine whether the cache has expired.
If the cache does not exist or has expired, the migration file is fetched from the remote source, saved to the cache, and the last updated time is updated.
If migration.yaml or metadata.json is corrupted (cannot be parsed as YAML or JSON) for any reason, the corrupted file is discarded and recreated.

## Expiration

The expiration period varies depending on the source of the migration file.

- github contents ref: `v?\d+\.\d+.\d+` => No expiration
- github contents ref: `\b[0-9a-f]{40}\b` => No expiration
- others: 3 days

## Ignoring the cache when running commands

```sh
yamledit run -no-cache
```

When `-no-cache` is specified, the cache is ignored, the migration file is fetched from the remote source, and the cache is updated.
