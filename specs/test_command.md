# test command

```sh
yamledit test [<migration> ...]
```

e.g.

```sh
yamledit test goreleaser-v2
```

Tests whether YAML files are modified as expected by ruleset files.
This command fails if any test files are not modified expectedly.

```
.yamledit/
  <ruleset>/
    ruleset.yaml
    <test>/
      test.yaml
      result.yaml
    ...
```

Rulesets are located at `.yamledit/<ruleset>/ruleset.yaml`.
Each test case is a subdirectory of `.yamledit/<ruleset>/` containing `test.yaml` and `result.yaml`.
The ruleset is applied to `test.yaml`, and the result is verified against `result.yaml`.
test command doesn't change any files actually.
During testing, `.rules[].files` in ruleset files is ignored.
If a test directory has `test.yaml` without `result.yaml`, a warning is output and the test is skipped.
If no arguments are specified, all rulesets under `.yamledit` are tested. If no test directories exist, a warning is output and the ruleset is skipped.
If test files aren't modified expectedly, diffs are outputted.

Unlike the `run` command, the `@` prefix is not required.
Also, `<migration>` does not support remote migration files such as URLs or GitHub Contents.
