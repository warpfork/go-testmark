testexec selfexercise file
==========================

We will run a script with some files provisioned:

[testmark]:# (whee/fs/a)
```
body-a
```

The script is simple:

[testmark]:# (whee/script)
```
echo hello
ls
```

The output is unsurprising:

[testmark]:# (whee/output)
```
hello
a
```

---

We can make subtests which add more files:

[testmark]:# (whee/then-more-files/fs/b)
```
body-b
```

[testmark]:# (whee/then-more-files/script)
```
ls
```

[testmark]:# (whee/then-more-files/output)
```
a
b
```

---

We can also affect the filesystem in subtests:

[testmark]:# (whee/then-touching-files/script)
```
echo "body-c" > ./c
ls
```

Notice that this doesn't inherit the file "b" from the last test --
this is because they're siblings, so they each got a copy of the *parent* filesystem state.

[testmark]:# (whee/then-touching-files/output)
```
a
c
```

---

If we make a subtest of *that* subtest, it also inherits a copy of those effects:

[testmark]:# (whee/then-touching-files/then-subtesting-again/script)
```
ls
cat ./c
```

[testmark]:# (whee/then-touching-files/then-subtesting-again/output)
```
a
c
body-c
```

---

Stdin can also be emulated:

[testmark]:# (using-stdin/input)
```
this is stdin and will be echoed
```


[testmark]:# (using-stdin/script)
```
cat - | sed 's/ is/ was/'  | sed s/will/should/
```

[testmark]:# (using-stdin/output)
```
this was stdin and should be echoed
```

---

The default recursion function won't reach everything if you don't formulate your paths correctly.

[testmark]:# (bad/script)
```
  # no-op
```

[testmark]:# (bad/then-missing-script/then-another-thing/script)
```
  # no-op, this will not be reached because it missing executable blocks in parent steps
```

[testmark]:# (bad/not-a-then-statement/script)
```
  # no-op, this script will not run because the subdirectory does not begin with "then-"
```

