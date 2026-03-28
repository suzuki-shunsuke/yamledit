# run command

```sh
yamledit run [@<migration> ...] [<yaml file>...]
```

Edits YAML files.
Reads migration files from `.yamledit/*.yaml` and applies migrations according to the configuration.
About migration files, please see [migration_file.md](migration_file.md)

Arguments starting with `@` are treated as migrations, otherwise as YAML files to edit.
If an argument starting with `@` has a suffix matching `\.ya?ml`, it is treated as a file path; otherwise it is treated as `.yamledit/<migration>.yaml`.
When migrations are specified, only the specified migrations are applied.

@foo => .yamledit/foo.yaml
@foo.yaml => foo.yaml
@foo/bar => Not support. raise error
@foo/bar.yaml => foo/bar.yaml

If no `<yaml file>` arguments are given, all files matching `**/*.yml` and `**/*.yaml` are targeted.
