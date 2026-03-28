# new command

```sh
yamledit new <migration name>
```

new command creates a migration file using a default template if it doesn't exist.
If the migration file already exists, this command does nothing.

This command has only one required argument.
The migration name must match the regular expression `^[a-z0-9_-]+$`.

The migration file is created at `.yamledit/<migration name>.yaml`.

e.g.

```sh
: Create .yamledit/foo.yaml
yamledit new foo
```

If `-test` option is set, `new` command creates test files instead of migration files.
About test, please see [test command](test_command.md).
