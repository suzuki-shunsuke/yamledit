# test command

```sh
yamledit test [<migration> ...]
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

test directories and files are created by `yamledit new -test` command.

```sh
: Create test files
yamledit new -test <migration> <test>
```

e.g.

```sh
yamledit new -test goreleaser-v2 standard
```

```
.yamledit/
  goreleaser-v2.yaml
  goreleaser-v2_test/ # directory created
    standard.yaml # created
    standard_result.yaml # created
```

```sh
yamledit test goreleaser-v2
```
