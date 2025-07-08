# AnCLI – Project Context

## Vision  
**AnCLI** is a *CLI-first* flash-card platform that teaches real-world command-line skills with **spaced repetition**.  
Each card executes an actual shell command in a **rootless OCI container**, grades the outcome (regex match or exit-status), and reschedules the card with the **FSRS 4-parameter** algorithm.  
Default sandbox hardening:

- `cap-drop=ALL`
- `--network none`
- read-only root filesystem
- `tmpfs /tmp`
- automatic container cleanup on Ctrl-C or timeout

---

## Core Architecture  

| Layer                  | Choice / Notes                                                |
|------------------------|--------------------------------------------------------------|
| **Language**           | Go = 1.24.4                                                  |
| **UI**                 | Bubble Tea TUI → toggle off with `--no-tui` for headless     |
| **Persistence**        | SQLite via `modernc.org/sqlite` (pure Go, no CGO)            |
| **Scheduler**          | FSRS (4-parameter)                                           |
| **Sandbox abstraction**| `ancli/sandbox` Go interface with pluggable **drivers**      |
| **First-class driver** | **Podman** (rootless, daemonless, cross-platform)            |
| **Optional drivers**   | Docker Engine / Colima, Devcontainer, custom runners         |
| **Deck layout**        | `<deck>/deck.yaml` + `cards.csv` + POSIX hook scripts        |

> **Modularity goal:** Every sandbox driver lives in its own Go package and self-registers via `init()`, so new back-ends can be dropped in without touching core logic.
---

## MVP Roadmap  

0. **Repo scaffold & CI** (Go 1.24.4 + GitHub Actions)  
1. **FSRS engine** and **SQLite storage** with 100 % unit-test coverage  
2. **Rootless Podman sandbox runner** (Alpine base, 5 s timeout)  
3. **Headless review loop** (pure CLI)  
4. **Bubble Tea TUI wrapper** (respect `--no-tui`)  
5. **Deck-lint tool** and example decks  

### Post-MVP Enhancements  

- Docker/Colima driver for users already on Docker  
- Devcontainer driver for VS Code users  
- WebAssembly sandbox PoC  
- Cloud sync & shared-decks marketplace  

---

### Guiding Principles  

1. **Rootless by default** – zero-trust, no privileged daemon.  
2. **Plug-and-play back-ends** – swap Podman for Docker with a single config flag.  
3. **100 % reproducible** – every card runs in an isolated container to guarantee deterministic grading.  
4. **TUI optional** – accessible via SSH or in CI pipelines.  


## Development Journal  

### 2025-07-08  
- Set up github CLI (logged into my gh account via `gh auth login`)
- Added the remote repo: `git clone git@github.com:justinlyon12/AnCLI.git`
- Initaized the golang project: `go mod init github.com/justinlyon12/ancli`
- Created /ancli/docs to hold this context and simplied README.md to just include the vision.
- Added additional repo structure via:
    ```
    mkdir -p cmd/flashcli internal/{storage,scheduler,sandbox/podman}
    touch cmd/flashcli/main.go docs/{context.md,journal.md}
    ```
- Created placeholder main func in /ancli/cmd/flashcli


<!--  
Template for future entries:

### 2025-07-09  
Finished wiring FSRS unit tests; pinned `go-fsrs v0.3.2` to avoid CGO.  
Switched SQLite journal_mode to MEMORY after tmpfs WAL errors.  
Next: prototype rootless sandbox runner.  
-->