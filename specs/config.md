# Configuration file

- project config: `.yamledit/yamledit.yaml`
- global config: `${XDG_CONFIG_HOME}/yamledit/config.yaml`

`.yamledit/yamledit.yaml` is the project configuration file.

The global config is resolved in the following order of priority:

1. `$YAMLEDIT_GLOBAL_CONFIG`
1. `${XDG_CONFIG_HOME}/yamledit/config.yaml`
1. (Linux, macOS) `${HOME}/.config/yamledit/config.yaml`
1. (Windows) `LocalAppData\yamledit\config.yaml`

```yaml
reusable_rules:
  - name: goreleaser-v2
    import: https://example.com/goreleaser-v2
  - name: checkout-cred
    import: github.com/suzuki-shunsuke/yamledit/.yamledit/checkout-cred.yaml:v0.1.0
```

## reusable_rules

`reusable_rules` is a list of reusable rules.
`yamledit add` command adds reusable rules to this field.
When an migration is specified like `yamledit run @<migration>`, rules are resolved in the following order:

1. `.yamledit/<migration>/ruleset.yaml`
1. reusable rule whose name is `<migration>` in project config
1. reusable rule whose name is `<migration>` in global config
