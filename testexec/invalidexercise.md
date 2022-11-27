testexec invalid file
===

Shows examples of invalid testexec files

---

Blocks cannot test combined output and stdout

[testmark]:# (stdout-combo/script)
```
  echo "hi"
```
[testmark]:# (stdout-combo/output)
```
hi
```
[testmark]:# (stdout-combo/stdout)
```
hi
```

---

Similarly, blocks may not test combined output and stderr

[testmark]:# (stderr-combo/script)
```
echo "hi"
```
[testmark]:# (stderr-combo/output)
```
hi
```
[testmark]:# (stderr-combo/stderr)
```
```
