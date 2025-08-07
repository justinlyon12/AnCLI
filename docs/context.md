# AnCLI – Project Context

## Vision  
**AnCLI** is a *CLI-first* flash-card platform that teaches real-world command-line skills with **spaced repetition**.  
Each card executes an actual shell command in a **rootless OCI container**, lets the user grade themselves (`Again | Hard | Good | Easy`), and reschedules with the **FSRS 4-parameter** algorithm.  

**Default sandbox hardening:**
- `cap-drop=ALL`
- `--network none` (cards must explicity opt-in to networking; the TUI shows a red banner when enabled)
- read-only root filesystem
- `tmpfs /tmp`
- automatic cleanup on Ctrl-C or timeout

---

## Core Architecture  

| Layer                  | Choice                                                       | Notes
|------------------------|--------------------------------------------------------------|---------------|
| **Language**           | Go 1.24.4                                                    |Single static binary & first class OCI ecosytem                                  |
| **CLI/Config**         | Cobra + Viper                                                |Sub-commands (`review`, `optimize`, `deck`, etc.) & env/config handling.         |
| **UI**                 | Bubble Tea TUI (opt-out via `--no-tui`)                      |Mature TUI the integrates cleanly with cobra                                     |
| **Persistence**        | SQLite via `modernc.org/sqlite`                              |Pure Go, no CGO to keep build simple                                             |
| **Scheduler**          | FSRS (4-parameter)                                           |Current best schedular (https://github.com/open-spaced-repetition/fsrs4anki.git) |
| **Optimizer**          | Python FSRS Optimizer                                        |Future: Optional optimizer for the FSRS algorithm (link above) no currently implmeented in Go. Plan it ot implmeent later to export review history, run python optimizer, and re-import tuned parameters.|
| **Sandbox abstraction**| `interal/sandbox`                                            |Drivers self-register via `init()`                                               |
| **First-class driver** | Podman                                                       |Rootless, daemonless, cross-platform                                             |
| **Optional drivers**   | Docker Engine / Colima, Devcontainer, custom runners         |Swapablle through config flag                                                    |
| **Packaging**          | .ancli tarballs                                              |deck.yaml, cards.csv, optional assets/                                           |

> **Modularity goal:** Every sandbox driver lives in its own Go package and self-registers via `init()`, so new back-ends can be dropped in without touching core logic.
---
## CLI Surface (Cobra)
    ```
    ancli review         # headless loop (TUI appears unless --no-tui)
    ancli optimize       # future: run offline FSRS parameter fitting
    ancli deck lint      # validate deck structure & hooks
    ancli deck pack      # build .ancli tarball
    ancli deck install   # unpack to ~/.ancli/decks
    ```
All commands respect --config, env vars, and global flags (e.g., --sandbox=docker).
---

## MVP Roadmap  

0. **Repo scaffold & CI** (Go 1.24.4 + GitHub Actions)  
1. **FSRS engine** and **SQLite storage** with 100 % unit-test coverage  
2. **Rootless Podman sandbox runner** (Alpine base, 5 s timeout)  
3. **Headless review loop** (pure CLI)  
4. **Bubble Tea TUI wrapper** (respect `--no-tui`)  
5. **Deck-lint tool** and example decks  

### Post-MVP Enhancements  

- Docker driver for users already on Docker  
- Devcontainer driver for VS Code users    
- Cloud sync & shared-decks marketplace  
- Add optimizer

---

### Guiding Principles  

1. **Rootless by default** – zero-trust, no privileged daemon.  
2. **Plug-and-play back-ends** – swap Podman for Docker with a single config flag.  
3. **100 % reproducible** – every card runs in an isolated container to guarantee deterministic grading.  
4. **TUI optional** – accessible via SSH or in CI pipelines.  


## Development Journal  
<!--  
Template for future entries:

#### 2025-07-09  
Finished wiring FSRS unit tests; pinned `go-fsrs v0.3.2` to avoid CGO.  
Switched SQLite journal_mode to MEMORY after tmpfs WAL errors.  
Next: prototype rootless sandbox runner.  
-->
### 2025 July (07)
#### 2025-07-08  
- Set up github CLI (logged into my gh account via `gh auth login`)
- Added the remote repo: `git clone git@github.com:justinlyon12/AnCLI.git`
- Initaized the golang project: `go mod init github.com/justinlyon12/ancli`
- Created /ancli/docs to hold this context and simplied README.md to just include the vision.
- Added additional repo structure via:
    ```
    mkdir -p cmd/flashcli internal/{storage,scheduler,sandbox/podman} docs bin
    touch cmd/flashcli/main.go docs/{context.md,journal.md}
    ```
- Created placeholder main func in /ancli/cmd/flashcli
- Created makefile with `help`, `build`, `test`, `lint`, `tidy` and targets for CI
- Added Github Actions CI:
    - made directory and CI file via:
        ```
        mkdir -p .github/workflows
        touch .github/workflows/ci.yml
        ```
    - added `make ci` to `/ancli/.github/workflows/ci.yml` to ensure that full test, lint, and build suite is run on every push or pull_request


<!--  
Template for future entries:

### 2025-07-09  
Finished wiring FSRS unit tests; pinned `go-fsrs v0.3.2` to avoid CGO.  
Switched SQLite journal_mode to MEMORY after tmpfs WAL errors.  
Next: prototype rootless sandbox runner.  
-->