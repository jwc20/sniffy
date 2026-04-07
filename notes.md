# notes

---

## Atomic saves

Vim, Emacs, and VScode uses `atomic saves`.

The save sequence is:

- `CREATE` sniffy.go~ — write new content to temp file
- `RENAME` sniffy.go — old file renamed away
- `CREATE` sniffy.go — temp file renamed into place
- `CHMOD` sniffy.go — permissions restored
- `REMOVE` sniffy.go~ — cleanup

## In-place saves

Fleet(Jetbrain) and other editors writes directly in file in-place:

The save sequence is:

- `CHMOD`
- `WRITE`

---