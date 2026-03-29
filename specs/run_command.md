# run command

```sh
yamledit run [@<migration> ...] [<yaml file>...]
```

Edits YAML files.
Reads migration files from `.yamledit/*.yaml` and applies migrations according to the configuration.
About migration files, please see [migration_file.md](migration_file.md)

Arguments starting with `@` are treated as migrations, otherwise as YAML files to edit.
When migrations are specified, only the specified migrations are applied.

`<migration>` is evaluated by the following order:

1. `@https://...`, `@http://...` => Fetch migration from URL
1. `@github.com/<owner>/<repo>/<path>[:<ref>]` => Fetch migration from URL
    1. `:<ref>` is optional. If it's empty, the default branch is used.
1. `@./github.com/...` => Local path `github.com/...`
1. `@foo.yaml` => foo.yaml
1. `@foo/bar.yaml` => foo/bar.yaml
1. `@foo` => .yamledit/foo.yaml
1. Other `@foo/bar` => Not support. raise error

If no `<yaml file>` arguments are given, all files matching `**/*.yml` and `**/*.yaml` are targeted.

When `<migration>` is a remote migration file such as a URL or GitHub Contents, it is [cached](cache.md).
