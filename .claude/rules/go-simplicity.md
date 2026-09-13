---
name: go-simplicity
description: Boring idiomatic Go — no helper/constructor/builder slop, straightforward control flow, stdlib first.
applies_to: "**/*.go"
---

# Go Simplicity

Code should look like an experienced Go engineer wrote it: simple, explicit, strongly typed, boring, easy to trace, minimal abstraction, minimal magic. Do not optimize for cleverness.

## No pointless helper extraction

Do not create a helper for every 2–3 lines. Be suspicious of `getDomainID`, `extractProjectID`, `resolveName`, `buildKey`, `parseValue`, `stringFrom`, `domainFrom` when the helper:

- has one caller
- only accesses a field
- only checks nil
- only calls `strings.TrimSpace`
- only calls `.String()`
- only performs a type assertion
- only forwards arguments

Bad:

```go
func domainIDFrom(record *DNSRecord) string {
	if record == nil {
		return ""
	}

	return record.DomainID.String()
}
```

Prefer `record.DomainID.String()` when `record` is already known to exist.

A helper should represent a real reusable concept, not hide straightforward code.

## Constructors, builders, options, factories

Constructors should establish invariants or hide meaningful setup. `NewProjectService(repo)` is fine as a stable construction API; `NewDomain(name)` returning `Domain{Name: name}` is not — write `Domain{Name: name}`.

No Java-style builders:

```go
// bad
deployment := NewDeploymentBuilder().WithProjectID(id).WithRegion(r).Build()

// good
deployment := Deployment{ProjectID: id, Region: r}
```

No functional options when all fields are required — `NewService(repo, logger, metrics)`. Options are for genuinely optional configuration.

No factories when the application has one concrete implementation. Construct the concrete dependency directly.

## Straightforward control flow, early returns

Bad:

```go
if err == nil {
	if project != nil {
		if project.Enabled {
			// 80 lines
		}
	}
}
```

Prefer:

```go
if err != nil {
	return err
}

if project == nil {
	return ErrProjectNotFound
}

if !project.Enabled {
	return nil
}

// main logic
```

Or, when simple, a single boolean expression. Keep the happy path visually obvious. Choose whichever is easier to read; do not optimize for clever one-liners.

## Remove useless temporaries and noise

```go
// bad
rawDomain := event.Domain
normalizedDomain := strings.TrimSpace(rawDomain)
domain := strings.ToLower(normalizedDomain)

// good
domain := strings.ToLower(strings.TrimSpace(event.Domain))
```

Intermediate variables should represent meaningful concepts — but do not compress so aggressively that readability decreases.

Clean up `if enabled == true` / `if enabled == false` → `if enabled` / `if !enabled`.

Prefer `var items []Item` over `make([]Item, 0)` unless capacity preallocation is justified. Do not create `make(map[string]string)` until the map needs writes.

Do not defensively copy slices/maps with no mutation threat:

```go
result := make([]string, len(input))
copy(result, input)
return result
```

unless ownership/mutation semantics require it.

## Signatures and return values

Pass the data the function actually needs, not giant config objects:

```go
// bad
func Process(ctx context.Context, cfg Config, options Options, metadata Metadata)
```

No option struct for a single parameter — `GetProject(ctx, id)`, not `GetProject(ctx, GetProjectOptions{ID: id})`.

No wrapper return structs — `func Exists(...) (bool, error)`, not `ExistsResult{Exists bool}`.

No needless named returns — `func GetProject(id string) (*Project, error)`. Avoid naked returns in non-trivial functions.

## Use the standard library

Before keeping a helper, check whether Go already has it: `strings.TrimSpace`, `strings.ToLower`, `slices.Contains`, `maps.Clone`, `errors.Is`, `errors.As`, `cmp.Or`, `strconv.Atoi`.

Avoid regex where `strings` functions suffice — regex should solve regex-shaped problems.

Avoid `fmt.Sprintf("%s", value)` when `value` is already a string; prefer `strconv.Itoa(n)` over `fmt.Sprintf("%d", n)` in simple paths. Do not rewrite readable formatting for micro-performance.

## Constants and switches

Constants should communicate meaning — `const duplicateKeyCode = 11000` is good; `const emptyString = ""` is not.

Prefer a readable `switch` over a dispatch map or a strategy interface when the set of cases is static:

```go
switch event.Type {
case "insert":
	return handleInsert(event)

case "delete":
	return handleDelete(event)

default:
	return ErrUnsupportedEvent
}
```

Use dispatch maps only when dynamic registration is genuinely valuable. Go does not require every branch to become polymorphism.

## Remove obvious comments

Delete `// Check if project exists.`, `// Return the result.`, `// Convert string to lowercase.`

Keep comments for business rules, invariants, non-obvious decisions, external system quirks, workarounds, and concurrency reasoning. Comments explain **why**, not syntax.

## Reduce concepts, not lines

Code should be shorter because unnecessary concepts disappeared, not because everything was compressed:

```go
// bad "cleanup"
if err := func() error { ... }(); err != nil { return fmt.Errorf(...) }
```

Prefer boring readable Go.
