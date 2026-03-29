# test command

```sh
yamledit test [<migration> ...]
```

e.g.

```sh
yamledit test goreleaser-v2
```

Tests whether YAML files are modified as expected by migration files.
This command fails if any test files are not modified expectedly.

```
.yamledit/
  <migration>.yaml
  <migration>_test/
    <test>.yaml
    <test>_result.yaml
    ...
```

Tests for `.yamledit/<migration>.yaml` are located at `.yamledit/<migration>_test/`.
The migration is applied to `.yamledit/<migration>_test/<test>.yaml`, and the result is verified against `.yamledit/<migration>_test/<test>_result.yaml`.
test command doesn't change any files actually.
During testing, `.rules[].files` in migration files is ignored.
Files ending with `_result.yaml` are excluded from test targets.
If a `<test>.yaml` exists without a corresponding `<test>_result.yaml`, a warning is output and `<test>.yaml` is skipped.
If no arguments are specified, all migration files under `.yamledit` are tested. If no test files exist, a warning is output and the migration is skipped.
If test files aren't modified expectedly, diff are outputted.

Unlike the `run` command, the `@` prefix is not required.
Also, `<migration>` does not support remote migration files such as URLs or GitHub Contents.
