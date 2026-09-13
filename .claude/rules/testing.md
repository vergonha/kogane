---
name: testing
description: Tests get the same cleanup — no mock-driven interfaces, no fixture frameworks.
applies_to: "**/*_test.go"
---

# Testing

## Do not create interfaces just to mock

Do not turn every internal type into an interface because "tests need mocks." Prefer testing real behavior where practical; use small interfaces around expensive or external boundaries only.

## Keep tests simple

Apply the same cleanup rules to test code. Remove:

- enormous test builders
- generic fixture systems
- helper pyramids
- mocks for pure logic
- repeated `any`
- excessive test abstractions

Prefer explicit table-driven tests where appropriate:

```go
tests := []struct {
	name string
	in   string
	want string
}{
	{"lowercase", "EXAMPLE.COM", "example.com"},
	{"trim", " example.com ", "example.com"},
}
```

Tests should not be harder to understand than the code they test.
