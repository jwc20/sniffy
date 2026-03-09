# plan

## 1. requirements

- Create auto-test application
- should have cli
- should have tui using bubbletea
- should be able to pick and choose which tests to run
- should be able to run tests in other languages, not just Go
- should be able to configure globally and per-project using yaml or toml files
  - should be able to create a file called `sniffa.yml` to configure the application (which directories to include, which files to ignore, configure languages, set history limit, set date format, etc)
- should be able to store state in `.sniffa.json`
- should have fuzzy-finder

## 2. data

- State
- state history

## 3. high-level design

- cli cmds (commands to ignore, add to queue, etc. Not to show running tests, or to run tests)
- tui app (bubbletea)
- file watcher (sniffa.go)
- file to store state
- message queue (maybe have a channel for file watcher to send messages to tui app)
