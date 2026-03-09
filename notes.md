# notes

---

## Atomic saves

Vim, Emacs, and VScode uses `atomic saves`.

The save sequence is:

- `CREATE` sniffa.go~ — write new content to temp file
- `RENAME` sniffa.go — old file renamed away
- `CREATE` sniffa.go — temp file renamed into place
- `CHMOD` sniffa.go — permissions restored
- `REMOVE` sniffa.go~ — cleanup

## In-place saves

Fleet(Jetbrain) and other editors writes directly in file in-place:

The save sequence is:

- `CHMOD`
- `WRITE`

---