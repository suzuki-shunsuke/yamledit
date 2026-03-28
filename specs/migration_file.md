# Migration Files

```yaml
rules:
  - path: "$"
    actions:
      - type: remove_keys
        keys:
          - age
```

- `.rules`: You can define multiple rules. Rules are evaluated in order from top to bottom, and the YAML file is edited accordingly.
- `.rules[].path`: Specifies the target YAML node using a YAML path (e.g. `"$"`, `"$.foo"`).
- `.rules[].actions`: You can define multiple actions. Actions are evaluated in order from top to bottom, and the YAML file is edited accordingly.
- `.rules[].actions[].type`: The action type. Configuration fields vary depending on the type.
  - Supported types: `remove_keys`, `rename_key`, `set_key`, `add_values`, `sort_key`, `remove_values`
  - `remove_keys`: Remove keys from maps
    - `.rules[].actions[].keys`: List of keys to remove
  - `rename_key`: Rename keys of maps
    - `.rules[].actions[].key`: The key to rename
    - `.rules[].actions[].new_key`: New key name
    - `.rules[].actions[].when_duplicate`: One of `skip` (default), `ignore_existing_key`, `remove_old_key`, `fail`
  - `set_key`: Set a value for a map key. You can specify behavior for when the key exists or does not exist.
    - `.rules[].actions[].key`: The key to set the value for
    - `.rules[].actions[].value`: The value to set for the key
    - `.rules[].actions[].skip_if_key_not_found`: boolean. If true, do nothing when the key does not exist
    - `.rules[].actions[].skip_if_key_found`: boolean. If true, do nothing when the key already exists
    - `.rules[].actions[].clear_comment`: boolean. If true, remove the comment on the existing key
    - `.rules[].actions[].insert_at`: Specifies where to insert a new key. Multiple conditions can be set. The first matching condition is used. If no condition matches, the key is appended at the end.
    - `.rules[].actions[].insert_at[].after_key`: Insert after the specified key
    - `.rules[].actions[].insert_at[].before_key`: Insert before the specified key
    - `.rules[].actions[].insert_at[].first`: Insert at the beginning
  - `add_values`: Add elements to lists
    - `.rules[].actions[].values`: List of values to add to the list
    - `.rules[].actions[].index`: The array index indicating where to insert values. 0 for the beginning. Negative values insert at (array length + index). In other words, -1 appends to the end of the array. Default is -1.
  - `sort_key`: Sort map keys
    - `.rules[].actions[].expr`: The sorting logic expression
  - `remove_values`: Remove elements from lists
    - `.rules[].actions[].expr`: The expression to select elements to remove
