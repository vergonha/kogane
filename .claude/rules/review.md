---
name: review
description: How to run the cleanup / review a PR — search patterns, the de-slop test, what must NOT be removed.
applies_to: "**"
---

# Review Process

Use this file when auditing a PR, reviewing a diff, or running a cleanup pass.

## Search patterns to audit

```txt
any                  interface{}          map[string]any
map[string]interface{}                    reflect.
recover(             panic(               fmt.Sprintf
strconv.             errors.New           fmt.Errorf
errors.As            errors.Is            == nil
!= nil               type .* interface    Factory
Builder              Manager              Processor
Helper               Utils                Mapper
Converter            Validator            Options
Params               DTO                  Entity
Model                ValueOr              Safe
Normalize            Extract              Resolve
Parse                GetString            AsString
ToString             With                 context.Background
context.TODO         sync.Pool            go func
make([]
```

These are places to **inspect**, not matches to change automatically.

## The de-slop test

Before adding or keeping code, ask:

- Does this handle something that can genuinely happen? If no, delete it.
- Is this complexity caused by poor typing upstream? If yes, fix the upstream type.
- Does this interface have more than one meaningful implementation or a consumer-driven purpose? If no, use the concrete type.
- Does this helper express a real concept? If no, inline it.
- Does this abstraction reduce total complexity? If no, remove it.
- Would plain Go be easier to understand? If yes, use plain Go.

## Refactoring process

Work feature-by-feature.

1. **Identify where data enters** — HTTP, queue, Kafka/NATS/RabbitMQ, MongoDB, PostgreSQL, Redis, webhook, external API, environment/config.
2. **Give the input a concrete type.**
3. **Validate/decode once.**
4. **Follow the data through the application**, removing unnecessary `any`, type assertions, generic maps, nil guards, converters, extractors, wrapper structs, interfaces, pass-through methods, duplicate DTOs, and generic helpers.
5. **Collapse forwarding layers.**
6. **Delete dead abstractions.**
7. **Run tests and static analysis:**

```bash
go test ./...
go vet ./...
staticcheck ./...
gofmt -l .
```

## Desired style

```go
func (s *Service) MapDomain(ctx context.Context, event DomainMapEvent) error {
	domain := strings.ToLower(strings.TrimSpace(event.Domain))

	return s.domains.Map(ctx, event.ProjectID, domain)
}
```

not:

```go
func (s *Service) MapDomain(ctx context.Context, raw any) error {
	event, err := convertToDomainMapEvent(raw)
	if err != nil {
		return fmt.Errorf("failed converting domain event: %w", err)
	}

	projectID := safeString(event.ProjectID)
	if projectID == "" {
		return nil
	}

	domain := normalizeStringValue(event.Domain)
	if domain == "" {
		return nil
	}

	return s.domainManager.ProcessDomainMapping(
		ctx,
		NewDomainMappingParams(projectID, domain),
	)
}
```

## Do NOT remove idiomatic Go

This is a complexity cleanup, not reckless deletion. These are correct and must stay when the failure or absence they handle is genuinely possible:

```go
if err != nil {
	return err
}

value, ok := m[key]

if project == nil {
	...
}
```

Also preserve: useful interfaces, meaningful error wrapping, boundary validation, context propagation, resource cleanup (`defer rows.Close()`, `defer resp.Body.Close()`), transaction rollback safety, mutexes protecting real shared state, needed channel synchronization, driver-specific error handling, correct integer/error checks, and security-related checks.

## Do not replace one kind of slop with another

Do not turn:

```go
value := data["domain"].(string)
```

into:

```go
value, ok := data["domain"]
if !ok {
	return ""
}

domain, ok := value.(string)
if !ok {
	return ""
}
```

and call that a cleanup. The correct fix is a typed struct. Likewise, never replace a direct call with interface → implementation → manager → helper → converter → validator → actual function.

## Outcome to aim for

More concrete structs; fewer `any` values, generic maps, type assertions, conversion helpers, tiny wrappers, pointless interfaces, forwarding layers, factories/builders/managers; less reflection, defensive nil handling, and silent fallback; simpler error handling; fewer redundant DTOs; more direct calls; narrower signatures; explicit business logic; validation concentrated at boundaries.

The metric is not files changed. It is:

> **Can an engineer trace the behavior without jumping through unnecessary abstractions?**

## review output

the review only exists once it has been posted to the PR. the final
assistant response is not the review.

there must always be exactly one PR-level comment. this is not optional.

inline comments are for specific problems in specific lines. post one per
finding, attached to the smallest relevant changed line: what is wrong +
why it matters + what to do about it.

the PR-level comment:

- if there were findings, it says in a couple of lines what the diff does
  and what came out of the review. it does not repeat the inline comments
  and it does not list everything that was checked.
- if there were no findings, its entire body is:

lgtm

nothing else. no summary of the diff, no list of what was checked, no
praise.

follow `writing.md`: lowercase, no emojis, no filler, no corporate review
language. every line carries information or it does not go in.

do not manufacture findings to make the review look useful.
