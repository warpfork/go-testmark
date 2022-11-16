testexec strict file
===

Shows examples of poorly formed testexec files. These will be skipped if strict mode is disabled.

---

A hunk with an empty intermediate step will not execute.

[testmark]:# (parent/script)
```
  # no-op
```

[testmark]:# (parent/then-missing-script/then-another-thing/script)
```
  # no-op, this will not be reached because it missing executable blocks in parent steps
```

---

Recursion won't happen if child blocks don't begin with `then-`
Any unrecognized pattern will be an error in strict mode.

[testmark]:# (norecurse/script)
```
  # no-op
```
[testmark]:# (norecurse/not-a-then-statement/script)
```
  # no-op, this script will not run because the subdirectory does not begin with "then-"
```

---

Blocks can't contain both a `script` and a `sequence` child

[testmark]:# (multiexec/script)
```
  # no-op, this script will not run because it has both types of exec blocks
```

[testmark]:# (multiexec/sequence)
```
  # no-op, this script will not run because it has both types of exec blocks
```

