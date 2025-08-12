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
- configurable resource guards per card: time, CPU/mem caps
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
- All commands respect --config, env vars, and global flags (e.g., --sandbox=docker).
---

## MVP Roadmap  

0. **Repo scaffold & CI** (Go 1.24.4 + GitHub Actions)  
1. **FSRS engine** and **SQLite storage** with 100 % unit-test coverage  
2. **Rootless Podman sandbox runner** (Alpine base, 5 s timeout)  
3. **Headless review loop** (pure CLI) 
4. **Deck-lint tool** and example decks   
5. **Bubble Tea TUI wrapper** (respect `--no-tui`)  


### Post-MVP Enhancements  

- Docker driver for users already on Docker  
- Devcontainer driver for VS Code users
- Lab runner driver for small multi-container topologies   
- Cloud sync & shared-decks marketplace  
- Add optimizer

---

### Guiding Principles  

1. **Rootless by default** – zero-trust, no privileged daemon.  
2. **Plug-and-play back-ends** – swap Podman for Docker with a single config flag.  
3. **Reproducible** – Cards run in an isolated container to provide deterministic behavior.  
4. **TUI optional** – accessible via SSH or in CI pipelines.  
5. **Determinism > realism** - prefer reproducable, single-container labs to fragile, complex topologies.

### Personas & High‑Value Decks (near‑term)
**Personas**: CLI beginners/upgraders; DevOps/SRE/sysadmins; network engineers; team leads/trainers; security‑minded orgs; content authors.  

**Potential Deck areas**:
- Linux ops drills (systemd, journald, files/ACLs) - default safe    
- Git for operators (rebase, bisect, conflict surgery) - default safe
- `jq`/`yq` pipelines for K8s manifests, AWS JSON, logs - default safe
- Kubernetes day‑2 (`kubectl` queries, rollouts, debug) - default safe
- Terraform basics with local providers - default safe
- Network automation 101 (`scrapli`/`netmiko` with recorded outputs) - default safe
- FRR/BIRD single‑node routing lab - requires `--network` and possibly lifted caps
- Vendor NOS syntax/show triage (XRd/cRPD/cEOS) - requires `--network` and possibly lifted caps & user‑supplied/licensed image

### UI/UX Decisions (v0.1)
- **Fast path to reps:** `ancli review` → queue opens immediately; time‑to‑first‑rep < 5 minutes.    
- **Safety ribbon:** top status shows safety level and why (e.g., net‑enabled due to `aws s3 ls`).
- **Minimal card chrome:** prompt, shell pane (stdout/stderr tail), rating row (`1–4` or `A/H/G/E`).
- **Helpful controls:** _hint_, _reset environment_, _view solution_, _open reference_ (hotkeys documented).
- **Interruptible:** `q` to pause/park card, `s` to skip/flag; resume exactly where left.
- **Author affordances:** `?` toggles card metadata (image digest, caps, env). Hidden by default.
- **Headless parity:** every TUI action has a CLI flag for SSH/CI use.
- **Failure messages:** human‑readable guidance (e.g., “Podman missing. Run: …”), not stack traces.

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
- Installed golangci-lint via `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- Verified could build via my Makefile and binary could be run (`./bin/flashcli`)

Tree: 
.
├── bin
│   └── flashcli
├── cmd
│   └── flashcli
│       └── main.go
├── docs
│   ├── context.md
│   └── journal.md
├── go.mod
├── internal
│   ├── sandbox
│   │   └── podman
│   ├── scheduler
│   └── storage
├── LICENSE
├── Makefile
└── README.md

### 2025 August (08)
#### 2025-08-06
- Implemented FSRS scheduler in `internal/scheduler/fsrs.go` using `go-fsrs/v3`
- Changed from auto-grading to manual `Again|Hard|Good|Easy` ratings per FSRS philosophy
- Created comprehensive unit tests for `fsrs.go` achieving 100% coverage
- Next: Demo & SQLite storage layer

#### 2025-08-07
- Added `examples/scheduler_demo.go` to demonstrate FSRS functionality
- Verified scheduler works correctly with different rating outcomes
- **Completed SQLite storage layer implementation:**
  - Created comprehensive data models (`internal/storage/models.go`)
  - Implemented database schema with auto-migration (`internal/storage/migrations.go`)
  - Built full CRUD operations for decks, cards, reviews, and assets (`internal/storage/sqlite.go`)
  - Added FSRS integration helpers for seamless conversion (`internal/storage/fsrs_helpers.go`)
  - Achieved 100% test coverage with comprehensive unit tests (`internal/storage/sqlite_test.go`)
  - Created storage + FSRS integration demo (`examples/storage_demo.go`)
- **Key architectural decisions:**
  - Used raw SQL queries for optimal performance
  - Auto-migration on database connection for zero-config setup
  - Embedded FSRS state in cards table for efficient queries
  - Symbolic linking approach for card prerequisites
  - Deck-level asset storage with filename-based referencing

**Outstanding TODOs for Storage Layer:**
    1. Prerequisite handling: JSON parsing and prerequisite card fetching logic
    2. JSON field validation: Proper parsing/validation for tags, environment_vars, capabilities, prerequisites
    3. Deck versioning: Implement deck update diff tracking and migration logic
    4. FSRS parameter customization: Per-deck FSRS parameter loading and application
    5. Asset extraction workflow: Integration with sandbox execution (extract assets to temp dirs)
    6. Review analytics: Additional queries for learning statistics and progress tracking
    7. Database optimization: Connection pooling and query performance tuning for large datasets

Qutstanding question: Given that AnCLI centers around executing commands, do we need a run_logs table in our DB to track things like stdout, stderr, exit, resources, etc.?
Next:Ready for Cobra CLI framework and Podman sandbox runner

#### 2025-08-09
- **Implemented sandbox abstraction layer with rootless Podman driver:**
  - Created hexagonal port/adapter architecture (`internal/sandbox/port.go`)
  - Built secure-by-default execution config with validation (`internal/sandbox/config.go`)
  - Implemented thread-safe driver registry with self-registration (`internal/sandbox/registry.go`)
  - Built Podman adapter with session-reuse lifecycle for performance (`internal/sandbox/podman/driver.go`)
  - Added comprehensive unit tests achieving 100% coverage on port interface
  - All tests pass including race detector (no race conditions detected)
- **Key architectural decisions documented in [ADR-001](ADR-001-sandbox-abstraction.md):**
  - Selected session-reuse container lifecycle for optimal performance/security balance
  - Implemented security hardening: cap-drop=ALL, RO filesystem, no network, tmpfs /tmp
  - Thread-safe container state management with mutex protection
  - Structured logging with correlation IDs for traceability
  - Error handling with wrapped errors and user-facing messages
- **Review loop architecture documented in [DESIGN-review-loop.md](DESIGN-review-loop.md):**
  - Hexagonal architecture with ReviewService coordinating Storage, Scheduler, and Sandbox ports
  - Session management with in-memory state tracking and proper cleanup
  - Comprehensive security model with container hardening and network opt-in
  - Configuration hierarchy and data flow documentation
- **Security features implemented:**
  - Rootless execution by default
  - All capabilities dropped (`--cap-drop=ALL`)
  - Read-only root filesystem with secure tmpfs mounts
  - Network disabled by default (explicit opt-in required)
  - No privilege escalation (`--security-opt=no-new-privileges`)
- **Performance optimizations:**
  - Container reuse reduces startup overhead from ~200ms to ~10ms per command
  - Running state verification (not just existence check)
  - Stdout/stderr separation to prevent container ID corruption
  - Context-based timeout handling with proper cancellation
- **Testing coverage:**
  - Port interface: 100% unit test coverage
  - Podman driver: Comprehensive unit tests including concurrency safety
  - Integration tests ready (conditional on Podman availability)
  - Race detector verification: clean (no race conditions)

- **Implemented Cobra CLI framework with headless review loop:**
  - Created new CLI structure in `cmd/ancli/` (renamed from flashcli)
  - Built comprehensive configuration system with Viper (`internal/config/config.go`)
    - Secure defaults with proper path expansion and environment variable support
    - Fixed environment variable binding for nested keys (ANCLI_DATABASE_PATH, etc.)
    - Database directory creation with 0700 permissions for security
  - Implemented review service port/adapter architecture (`internal/review/`)
    - Clean domain types in `internal/domain/types.go` for shared Rating/CardState/ExecutionResult
    - ReviewService port interface with session management 
    - Service implementation integrating storage, scheduler, and sandbox layers
    - Proper error handling with wrapped errors and user-facing messages
  - Built complete review command (`cmd/ancli/review.go`)
    - Interactive review session with card presentation and command execution
    - Integration with FSRS scheduling and SQLite storage
    - Secure sandbox execution with rootless Podman containers
    - User rating input (1-4 / Again/Hard/Good/Easy) with validation
    - Session statistics and cleanup
  - **All quality gates achieved:**
    - `make ci` passes: tests, linting, race detector, build
    - 100% test coverage maintained on existing domain/storage/sandbox layers
    - Hexagonal architecture principles preserved
    - Security hardening: rootless execution, capability drops, RO filesystem
    - Structured logging with correlation IDs
    - Error handling with %w wrapping and user-friendly messages
- **Key architectural decisions:**
  - Used domain types (`internal/domain/`) to avoid adapter dependencies in ports
  - Reused existing storage FSRS conversion methods (ToFSRSCard/UpdateFromFSRSCard)
  - Simple string-to-[]string command parsing with `strings.Fields()` 
  - In-memory session management (suitable for single-user CLI tool)
  - Graceful JSON parsing with ignored errors for non-critical fields
- **Ready for production use:** `ancli review` command fully functional
  - Connects all layers: storage, scheduler, sandbox, configuration
  - Can execute shell commands in secure containers
  - Updates FSRS scheduling based on user performance ratings
  - Records detailed review history and timing metrics

#### 2025-08-12
**Completed Deck Specification and Validation System:**
- **Comprehensive deck format specification**:
  - `deck.yaml` format with metadata, container config, cleanup behavior, FSRS params, learning settings
  - `cards.csv` format with all necessary fields: key, title, command, description, setup, cleanup, prerequisites, verify, hint, solution, explanation, difficulty, tags
  - Asset management system for supporting files
- **Complete example deck**: "Linux File Operations" with 20 cards demonstrating:
  - Command dependencies and state management through prerequisites
  - Progressive difficulty from basic (ls, mkdir) to advanced (tar, symlinks)
  - Self-verification patterns teaching users to check their own work
  - Comprehensive setup/cleanup for repeatable practice
  - Multiple solution alternatives and detailed explanations
- **Robust validation system** (`internal/deck/validator.go`):
  - Structure validation (required files, format checking)
  - Content validation (required fields, duplicate keys, circular dependencies)
  - Security warnings (dangerous commands, network usage)
  - Usability checks (missing hints/explanations, difficulty progression)
  - User-friendly error messages with specific line/column references
- **CLI integration** (`cmd/ancli/deck.go`):
  - `ancli deck lint` command with verbose output and JSON support
  - Clean integration with existing Cobra command structure
- **Complete documentation** (`docs/DECK_AUTHORING.md`):
  - Quick start guide for new deck authors
  - Comprehensive reference covering all features
  - Best practices and common patterns
  - Troubleshooting guide and validation error reference

**Key innovations in the specification**:
1. **Prerequisites for command dependencies**: Not learning order, but ensuring required state exists (e.g., directory must exist before cd'ing into it)
2. **Self-verification commands**: Teaching users how to check their own work rather than automated validation
3. **Progressive state management**: Setup/cleanup commands enable repeatable practice
4. **Security-first container configuration**: Secure defaults with explicit opt-in for elevated permissions
5. **Multiple solutions support**: Accommodates different valid approaches to the same task
6. **Comprehensive explanations**: Focus on understanding output, not just memorizing commands

**Validation Results**: Example deck passes all validation checks (0 errors, 0 warnings)

Next: Implement TUI with Bubble Tea.
