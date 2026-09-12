# Kogane — Engineering Rules

Go codebase. The goal is code that reads like an experienced Go engineer wrote it:
simple, explicit, strongly typed, boring, easy to trace, minimal abstraction, minimal magic.

The overriding principle:

> **Make untrusted input safe at the edge. Keep trusted Go code boring everywhere else.**

Apply these rules when writing code, reviewing PRs, and running cleanup passes.
Load the file relevant to what you are touching.

| Rules | Covers |
| --- | --- |
| [typing](.claude/rules/typing.md) | `any` / `interface{}` / generic maps, conversion helpers, pointers, generics, custom types, DTO & mapper duplication |
| [go-simplicity](.claude/rules/go-simplicity.md) | Pointless helpers, constructors/builders/options/factories, control flow, temporaries, stdlib first, comments |
| [architecture](.claude/rules/architecture.md) | Interfaces, forwarding layers, data access, transactions, DI, globals, reflection, packages |
| [errors](.claude/rules/errors.md) | Error handling & wrapping, nil handling, fallback-to-zero, custom error types, logging, panic/recover |
| [concurrency](.claude/rules/concurrency.md) | Goroutines, channels, `context.Context`, timeout policy |
| [boundaries](.claude/rules/boundaries.md) | Decoding JSON once, defensive type assertions, where validation belongs, explicit business rules |
| [testing](.claude/rules/testing.md) | Mocks, fixtures, table-driven tests |
| [review](.claude/rules/review.md) | Search patterns, the de-slop test, refactoring process, what must NOT be removed |
| [safety](.claude/rules/safety.md) | What a cleanup is allowed to modify |


@.claude/rules/typing.md
@.claude/rules/go-simplicity.md
@.claude/rules/architecture.md
@.claude/rules/errors.md
@.claude/rules/concurrency.md
@.claude/rules/boundaries.md
@.claude/rules/testing.md
@.claude/rules/review.md
@.claude/rules/safety.md
