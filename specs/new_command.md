# new command

```sh
yamledit new <migration name>
```

new command creates a ruleset file and test files using built-in templates if they don't exist.

The following files are created:

```
.yamledit/
  <migration>/
    ruleset.yaml # Ruleset file
    normal/
      test.yaml # Test file
      result.yaml # Expected Result
```

If files already exist, they aren't changed.
Each file is created independently — if the migration file already exists but test files don't, test files are still created.

This command has only one required argument.
The migration name must match the regular expression `^[a-z0-9_-]+$`.

About test files, please see [test command](test_command.md).
