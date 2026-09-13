---
name: concurrency
description: Goroutines, channels, context propagation, and timeout policy.
applies_to: "**/*.go"
---

# Concurrency and Context

## Do not add goroutines unnecessarily

Be suspicious of `go func() { ... }()` added merely to make something "non-blocking." Every goroutine introduces lifecycle, cancellation, race potential, error propagation, and shutdown complexity. Use concurrency when the operation actually benefits from it.

## No channel-based architecture for synchronous work

Bad:

```txt
handler → channel → worker → channel → processor
```

when the logic is simply `processor.Process(ctx, event)`.

Channels are synchronization primitives, not an architectural requirement. Do not use them to make normal function calls look concurrent.

## Use `context.Context` correctly

Do not:

- store context permanently on structs
- create `context.Background()` deep in request processing
- accept `context.Context` where cancellation/deadlines are irrelevant
- nil-check context (`if ctx == nil { ctx = context.Background() }`)

A context parameter should not be nil. Pass the caller's context through I/O boundaries:

```go
func (s *Service) GetProject(ctx context.Context, id string) (*Project, error)
```

Do not create new contexts just to satisfy a function signature.

## Timeouts belong at boundaries

Do not mechanically write:

```go
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
```

inside every repository/service function. Repeated nested arbitrary timeouts are difficult to reason about; timeout policy usually belongs at meaningful boundaries.

## Keep what is genuinely needed

Mutexes protecting real shared state and channel synchronization that is actually required must stay.
