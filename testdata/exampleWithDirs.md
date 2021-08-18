an example with "directories"
=============================

This testmark implementation provides convenience functions for treating
the data hunk names as if they were "paths" in some sort of "filesystem".

So, in this file we'll exercise that by using a bunch of hunk names that have slashes in them.

### directories can be implied

[testmark]:# (one/two)
```text
foo
```

[testmark]:# (one/three)
```text
bar
```

### a path can have both children and data

(Maybe avoid leaning on this, though, because some would argue it's weird.)

[testmark]:# (one)
```text
baz
```

### directories can be implied deeply

[testmark]:# (really/deep/dirs/wow)
```text
zot
```

### files in the same directory don't have to be subsequent

[testmark]:# (one/four/bang)
```text
mop
```
