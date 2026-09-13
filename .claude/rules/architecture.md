---
name: architecture
description: Interfaces, layering, packages, DI, globals, data-access — collapse abstraction that only forwards.
applies_to: "**/*.go"
---

# Architecture

## Interfaces need a real reason

Go interfaces are usually defined by the consumer and kept small. Do not create an interface because "services should have interfaces."

Be suspicious of:

```go
type ProjectService interface {
	CreateProject(...)
	GetProject(...)
	UpdateProject(...)
	DeleteProject(...)
	ListProjects(...)
	ValidateProject(...)
}
```

especially with one implementation. No `IFoo` / `FooImpl` Java architecture — prefer `type DomainManager struct{}`.

Use an interface only for: multiple implementations, substitution, testing at that boundary, plugin behavior, or a narrow consumer contract.

When justified, define only what the consumer needs:

```go
type projectGetter interface {
	Get(ctx context.Context, id string) (*Project, error)
}
```

Do not expose a giant `Repository` interface just because the concrete repository has many methods.

## Remove forwarding layers

Look for chains like handler → controller → service → manager → processor → repository → store → database client, where most layers just forward arguments.

```go
func (s *Service) GetProject(ctx context.Context, id string) (*Project, error) {
	return s.manager.GetProject(ctx, id)
}
```

If a layer contains no meaningful business logic, remove it. One meaningful layer beats five pass-through layers.

Same for pass-through wrappers whose entire body is `return dependency.Do(...)`, or:

```go
result, err := dependency.Do(...)
if err != nil {
	return nil, err
}
return result, nil
```

Simplify to `return dependency.Do(...)`.

## Data access

Use the minimum layering that matches the application. A repository around SQL/Mongo queries is reasonable. A repository wrapped by another object that forwards everything is not.

Keep query code obvious — plain SQL is usually easier to read than a query builder written for three static queries:

```go
const query = `
	SELECT id, name
	FROM projects
	WHERE id = $1
`
```

Do not invent `TransactionManager` / `UnitOfWork` / `TransactionProvider` / `TransactionRunner` when this suffices:

```go
tx, err := db.BeginTx(ctx, nil)
if err != nil {
	return err
}

defer tx.Rollback()

...

return tx.Commit()
```

## Dependency wiring

Normal Go DI is `service := NewService(repo, logger)`. Do not introduce containers, service locators, registries, providers, dependency graphs, or reflection-based injection. Explicit wiring is a strength of Go.

Globals are fine for true constants and immutable package state. Do not use mutable package globals (`defaultClient`, `globalConfig`, `singleton`) as a shortcut for dependency wiring.

## Avoid reflection

`reflect.` is suspicious in ordinary business logic. Prefer explicit typed logic; do not use reflection to avoid writing five obvious lines. It is reasonable in serializers, frameworks, generic libraries, and tooling — rarely in handlers/services/domain logic.

## No premature performance machinery

Do not introduce `sync.Pool`, manual buffer reuse, unsafe conversions, custom allocators, elaborate caches, goroutine pools, or lock-free structures without evidence. Simple correct code first; optimizations solve measured problems.

## Packages

Do not create a package per type or helper:

```txt
internal/
  domainparser/
  stringutils/
  validationhelper/
  projectmapper/
  pointerhelper/
  responsebuilder/
```

A package should represent a coherent domain or capability.

Audit `utils`, `helpers`, `common`, `shared`, `misc`, `core`, `base` — they accumulate unrelated abstractions. Move useful functions to the domain that owns them, delete trivial helpers, and do not create a new generic utility package during cleanup.

## "AI architecture" smell

Be particularly suspicious of this combination inside one small feature:

```txt
interfaces.go  factory.go  builder.go  mapper.go  converter.go
validator.go   utils.go    helpers.go  manager.go processor.go
service.go     repository.go
```

Determine what each file actually does. Collapse layers that merely forward calls or transform identical structures.
