# Design: Review Loop Architecture

## Overview

The review loop is the core of AnCLI's learning experience. It integrates FSRS scheduling, SQLite storage, and rootless sandbox execution into a cohesive interactive session where users execute commands and rate their performance.

## Architecture

```
                         ┌─────────────────┐
                         │   Review CLI    │
                         │  (cmd/ancli)    │
                         └─────────┬───────┘
                                   │
                                   ▼
                         ┌─────────────────┐
                         │  Review Service │
                         │(internal/review)│
                         └─────────┬───────┘
                                   │
              ┌────────────────────┼────────────────────┐
              │                    │                    │
              ▼                    ▼                    ▼
    ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
    │  Storage Port   │  │ Scheduler Port  │  │  Sandbox Port   │
    │ (SQLite Adapter)│  │ (FSRS Adapter)  │  │ (Podman Adapter)│
    └─────────────────┘  └─────────────────┘  └─────────────────┘
```

## Key Components

### ReviewService Port (`internal/review/port.go`)
```go
type ReviewService interface {
    StartSession(ctx context.Context, opts SessionOptions) (*Session, error)
    GetNextCard(ctx context.Context, sessionID string) (*ReviewCard, error) 
    SubmitReview(ctx context.Context, sessionID string, cardID int, rating Rating, result *ExecutionResult) error
    EndSession(ctx context.Context, sessionID string) (*SessionStats, error)
}
```

### Session Management
- **In-memory session tracking** - Simple map for single-user CLI tool
- **Session lifecycle** - Start → GetCard → Execute → Rate → Repeat → End
- **State isolation** - Each session maintains its own card queue and progress

### Card Execution Flow
1. **Card Selection** - Query storage for due cards, shuffle if requested
2. **Metadata Resolution** - Merge deck defaults with card-specific overrides
3. **Sandbox Execution** - Rootless container with security hardening
4. **Performance Rating** - User input validation and FSRS integration
5. **State Updates** - Persist FSRS scheduling and review history

## Security Model

### Container Hardening
- `--cap-drop=ALL` - Remove all Linux capabilities
- `--security-opt=no-new-privileges` - Prevent privilege escalation
- `--read-only` - Read-only root filesystem
- `--tmpfs /tmp` - Secure temporary filesystem
- `--network=none` - No network access by default

### Network Opt-in
- Cards must explicitly declare network requirement
- CLI shows security warning when network enabled
- User can override with `--no-network=false` flag

### Resource Controls
- Per-command timeouts (default 30s)
- Memory and CPU limits (configurable)
- Container reuse for performance with session cleanup

## Data Flow

### Review Session
```
User runs: ancli review --deck-id=1 --max-cards=10

1. Load Configuration (Viper)
   ├── CLI flags override env vars
   ├── Env vars override config file  
   └── Config file overrides defaults

2. Initialize Dependencies
   ├── SQLite connection + auto-migration
   ├── FSRS scheduler with default parameters
   └── Podman driver with rootless config

3. Start Session
   ├── Query cards (filtered by deck, due date, limits)
   ├── Shuffle if requested
   └── Create session state with card queue

4. Review Loop (for each card)
   ├── Display card metadata
   ├── Wait for user confirmation
   ├── Execute command in container
   ├── Show stdout/stderr output
   ├── Collect user rating (1-4)
   ├── Update FSRS scheduling
   ├── Record review in database
   └── Update session progress

5. End Session
   ├── Show statistics (cards reviewed, time spent)
   ├── Cleanup containers
   └── Remove session from memory
```

### Error Handling
- **Graceful degradation** - Allow rating even on command failure
- **User-friendly messages** - No stack traces in production
- **Context propagation** - Cancellation and timeouts throughout
- **Wrapped errors** - `fmt.Errorf(..., %w, err)` for error chains

## Configuration

### Hierarchy (highest to lowest precedence)
1. CLI flags (`--database-path=/tmp/test.db`)
2. Environment variables (`ANCLI_DATABASE_PATH=/tmp/test.db`) 
3. Config file (`~/.ancli/ancli.yaml`)
4. Built-in defaults

### Key Settings
```yaml
database:
  path: ~/.ancli/ancli.db

sandbox:
  driver: podman
  default_image: alpine:3.18
  default_timeout: 30s
  network_enabled: false

review:
  max_cards_per_session: 20
  session_timeout: 30m
  auto_advance: false

log_level: info
log_json: false
```

## Testing Strategy

### Unit Tests
- **Mock dependencies** - Storage, Scheduler, Sandbox interfaces
- **Behavior verification** - Session lifecycle, error handling
- **Edge cases** - Empty decks, invalid ratings, timeouts

### Integration Tests  
- **Database integration** - Real SQLite with test data
- **FSRS integration** - Verify scheduling calculations
- **Sandbox integration** - Conditional on Podman availability

### Golden Tests (Future)
- **CLI output consistency** - Capture and verify terminal output
- **Regression protection** - Detect unexpected UI changes

## Performance Considerations

### Container Reuse
- **Session lifecycle** - Reuse container across cards in single session
- **Performance gain** - ~200ms → ~10ms per command execution
- **Security balance** - Container isolated between sessions

### Database Optimization
- **Connection reuse** - Single connection per session
- **Prepared statements** - Reused queries for card/review operations
- **Transaction batching** - Future optimization for bulk operations

### Memory Management
- **Session cleanup** - Automatic cleanup on session end
- **Card queue limits** - Prevent unbounded memory growth
- **Context cancellation** - Proper resource cleanup on interrupts

## Future Enhancements

### TUI Integration
- Replace CLI prompts with Bubble Tea components
- Progress bars, keyboard shortcuts, visual card presentation
- All logic remains in ReviewService (no business logic in TUI)

### Advanced Features
- Multiple deck support with weighted selection
- Custom FSRS parameters per deck
- Detailed analytics and progress tracking
- Card prerequisites and learning paths

## Security Considerations

### Container Escape Prevention
- Rootless execution eliminates most privilege escalation paths
- Read-only filesystem prevents persistence attacks
- Network isolation prevents data exfiltration
- Capability dropping reduces attack surface

### Data Protection
- Database stored in user directory with 0700 permissions
- No sensitive data logged at info level
- Correlation IDs for audit trails without exposing content
- Secure cleanup of temporary files and containers

This design enables secure, performant, and maintainable command-line skill learning through spaced repetition and containerized execution.
