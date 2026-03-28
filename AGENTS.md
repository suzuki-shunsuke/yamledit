# AI Assistant Guidelines

Please read CONTRIBUTING.md first.

### Auto fix lint errors

Note that only a few errors can be fixed by this command.
Many lint errors needs to be fixed manually.

```sh
golangci-lint fmt
```

## Run yamledit locally

Show help:

```sh
go run ./cmd/yamledit help-all
```

## Debugging

Enable debug logging:

```sh
export YAMLEDIT_LOG_LEVEL=debug
```
