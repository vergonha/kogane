---
name: errors
description: Error handling, nil handling, fallback-to-zero, custom error types, logging, panic/recover.
applies_to: "**/*.go"
---

# Errors, Nil, and Logging

## No fallback-to-zero-value programming

Search for code that silently converts invalid states into `""`, `0`, `false`, `nil`, `[]T{}`, `map[K]V{}`:

```go
if value == nil {
	return ""
}

if err != nil {
	return nil
}

if project == nil {
	return &Project{}
}
```

Do not hide invalid states. If data is required, return an error.

```go
// bad
func projectID(project *Project) string {
	if project == nil {
		return ""
	}

	return project.ID
}
```

Prefer fixing the caller so `project` cannot be nil. If absence is genuinely possible, `return ErrProjectNotFound`. Validate/narrow once, then continue with clean code.

## Nil checks must match real nullable states

Bad:

```go
func handleProject(project *Project) error {
	if project == nil {
		return errors.New("project is nil")
	}

	if project.Config == nil {
		return errors.New("project config is nil")
	}

	if project.Config.Domain == nil {
		return errors.New("domain is nil")
	}

	// actual logic
}
```

If those fields are required by the application model, redesign the types (see `typing.md`) so the checks disappear.

## Simplify redundant error handling

```go
// bad
result, err := repo.Get(ctx, id)
if err != nil {
	return nil, err
}

return result, nil

// good
return repo.Get(ctx, id)
```

Do not wrap with no added context:

```go
return fmt.Errorf("error: %w", err) // adds nothing
```

Wrap only where it materially improves debugging:

```go
return fmt.Errorf("load project %s: %w", id, err)
```

An error should never become:

```txt
failed to process project: failed to get project: failed to retrieve project:
failed to query project: sql: no rows
```

## Do not swallow errors

Inspect `if err != nil { return nil }`, `log.Println(err); return nil`, and `_ = doSomething()`. Decide whether ignoring is intentional; do not turn bugs into silent success. When ignoring is correct, make the reasoning obvious:

```go
if err := cache.Delete(ctx, key); err != nil {
	logger.Warn("failed to invalidate cache", "key", key, "error", err)
}
```

only when cache invalidation failure genuinely should not fail the operation.

## Keep error types minimal

Do not build hierarchies (`ValidationError`, `RepositoryError`, `ServiceError`, `DomainError`, `InternalError`) unless callers actually behave differently.

Prefer:

```go
var ErrProjectNotFound = errors.New("project not found")
```

with `errors.Is(err, ErrProjectNotFound)`. Use structured custom errors only when they carry meaningful information.

## Use `errors.Is` / `errors.As` idiomatically

Never inspect error strings:

```go
if strings.Contains(err.Error(), "duplicate") { // bad
```

Prefer the driver's supported error type/code:

```go
var writeErr mongo.WriteException
if errors.As(err, &writeErr) {
	...
}
```

Keep it proportional — do not turn a known-driver check into generic error inspection machinery.

## Logging

Do not narrate execution:

```go
logger.Info("entering CreateProject")
logger.Info("validating project")
logger.Info("calling repository")
```

Log meaningful events: failures, important state transitions, operational decisions, external interactions worth tracing.

Do not log and return the same error at every layer. Libraries/services return errors; handlers/workers/supervisors decide when to log. Log once, at the boundary responsible for handling the failure.

## `panic` and `recover`

Do not use `panic` for validation failures, missing DB rows, network errors, or malformed input — return errors. Panic is for states that make initialization or continued execution impossible.

Do not use `recover()` to turn programming bugs into control flow, and never inside normal business functions. Recover only at genuine process/request boundaries where keeping the process alive is intentional.
