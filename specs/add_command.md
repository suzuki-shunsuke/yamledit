# add command

Adds a Reusable Rule to the configuration file.

```sh
yamledit add [-global (-g)] [--force (-f)] [<name>] [<migration>]
```

When `-g` is set, the rule is added to the global config.

1. Fails if `<migration>` does not have a prefix of `https://`, `http://`, or `github.com/`.
1. Fails if `<name>` does not match `^[a-z0-9_-]+$`.
1. Fails if `.yamledit/yamledit.yaml` exists and the key `<name>` already exists in `names` of `.yamledit/yamledit.yaml`.
    1. When `-f` is set, the existing entry is overwritten.
1. Downloads `<migration>` and saves it to the cache. See [cache.md](cache.md) for details.
1. Creates `.yamledit` and `.yamledit/yamledit.yaml` if they don't exist.
1. Adds the name to `reusable_rules` in `.yamledit/yamledit.yaml`. go-yamledit can be used for YAML editing.

```yaml
reusable_rules:
  - name: "<name>"
    import: "<migration>"
```
