# ADR-001: Sandbox Abstraction Layer

## Status
Accepted

## Context
AnCLI requires secure execution of arbitrary shell commands for flashcard practice. The system must provide strong isolation between the host and container while maintaining good performance for interactive learning sessions.

## Decision
We will implement a hexagonal architecture with a sandbox port/adapter pattern that supports multiple container backends through driver self-registration.

### Architecture Components

1. **Port Interface** (`internal/sandbox/port.go`)
   - `Sandbox` interface defines `Run()`, `Cleanup()`, and `Name()` methods
   - `ExecutionConfig` struct encapsulates all execution parameters
   - `ExecutionResult` provides structured output with timing and metadata

2. **Driver Registry** (`internal/sandbox/registry.go`)
   - Global registry with thread-safe driver registration
   - Self-registration pattern via `init()` functions
   - Factory pattern for driver instantiation

3. **Podman Adapter** (`internal/sandbox/podman/`)
   - First-class implementation using rootless Podman
   - Session-reuse lifecycle for performance optimization
   - Comprehensive security hardening

### Container Lifecycle Strategy

We selected **session-reuse** as the default lifecycle:

- **PerCard**: New container per command (high isolation, ~200ms overhead)
- **SessionReuse**: Reuse container across cards (balanced approach) ✅ **SELECTED**
- **DeckPersistent**: Long-running container per deck (fast, lower isolation)

**Rationale**: Session-reuse provides the optimal balance of security isolation (host-container boundary) and performance (eliminates startup overhead between cards within a session).

### Security-by-Default Configuration

```go
// Default security posture
args := []string{
    "--cap-drop=ALL",                    // No capabilities
    "--security-opt=no-new-privileges",  // Prevent escalation
    "--read-only",                       // Immutable root FS
    "--tmpfs", "/tmp:rw,noexec,nosuid",  // Secure temp space
    "--network=none",                    // No network by default
}
```

## Consequences

### Positive
- **Modularity**: New drivers can be added without touching core logic
- **Security**: Consistent security defaults across all drivers
- **Performance**: Session-reuse eliminates container startup latency
- **Testability**: Port interface enables easy mocking and testing
- **Observability**: Structured logging with correlation IDs built-in

### Negative
- **Complexity**: More complex than direct Podman integration
- **State Management**: Session containers require careful lifecycle management
- **Resource Usage**: Long-running containers consume memory even when idle

### Risks & Mitigations
- **Race Conditions**: Mitigated with mutex protection around container state
- **Container Leaks**: Mitigated with explicit cleanup and timeout handling  
- **Security Bypass**: Mitigated with whitelist approach and default-deny posture

## Implementation Details

### Thread Safety
- Driver state protected by `sync.Mutex`
- Container ID and name access serialized
- Registry operations are thread-safe

### Error Handling
- Validation at config level with clear error messages
- Wrapped errors with `%w` for error chain analysis
- No panics in library code

### Performance Optimization
- Container reuse reduces startup overhead from ~200ms to ~10ms per command
- Podman inspection checks running state, not just existence
- Stdout/stderr separation prevents ID corruption from warnings

## Alternatives Considered

1. **Direct Podman Integration**: Simpler but less flexible, harder to test
2. **Docker-First**: Ruled out due to daemon requirement and security concerns
3. **Per-Command Containers**: Too slow for interactive learning experience

## References
- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- [Podman Security Guide](https://docs.podman.io/en/latest/markdown/podman-run.1.html#security-options)
- [Container Lifecycle Patterns](https://12factor.net/disposability)
