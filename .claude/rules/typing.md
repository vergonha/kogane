---
name: typing
description: Concrete types over any/interface{}/generic maps; pointers, custom types, DTO duplication.
applies_to: "**/*.go"
---

# Typing

> Validate untrusted data at the boundary. Use concrete types everywhere else.

## Remove unnecessary `any` / `interface{}`

Audit every `any`, `interface{}`, `map[string]any`, `map[string]interface{}` and ask why the value is untyped.

Bad:

```go
func getDomain(data map[string]any) string {
	value, ok := data["domain"]
	if !ok {
		return ""
	}

	domain, ok := value.(string)
	if !ok {
		return ""
	}

	return domain
}
```

Prefer:

```go
type DomainEvent struct {
	Domain string `json:"domain"`
}
```

Then `event.Domain`.

Do not carry generic maps through the application and repeatedly recover types from them. If the JSON shape is known, unmarshal directly into a struct.

## No generic conversion helpers

Be suspicious of `toString`, `asString`, `stringValue`, `safeString`, `toInt`, `asInt`, `toBool`, `toMap`, `asMap`, `getString`, `getOptionalString`, `valueOrDefault`.

Bad:

```go
func toString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case fmt.Stringer:
		return v.String()
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}
```

Ask instead: what type is this value actually supposed to be?

```go
func normalizeDomain(domain string) string {
	return strings.ToLower(strings.TrimSpace(domain))
}
```

Do not accept `any` just to make a helper "flexible."

## No generic map accessors

```go
func GetString(data map[string]any, key string) string
func GetInt(data map[string]any, key string) int
func GetBool(data map[string]any, key string) bool
```

This is evidence the data should have been decoded into a struct. Define the struct and use its fields.

## No `map[string]any` for database models

If MongoDB, JSONB, Redis, or another store returns known document shapes, define structs.

Bad:

```go
var document map[string]any
id, ok := document["domain"].(primitive.ObjectID)
```

Prefer:

```go
type DNSDocument struct {
	Domain primitive.ObjectID `bson:"domain"`
}
```

Fix broad types at the data-access boundary rather than adding extractors everywhere.

## Stop using pointers for everything

Review `*string`, `*bool`, `*int`, `*time.Time` struct fields. Do not use pointers merely to distinguish "missing" from zero unless the distinction actually matters.

Bad:

```go
type Config struct {
	Enabled *bool
	Port    *int
	Name    *string
}
```

If required, use values. Pointers should communicate real optionality or identity/mutation semantics.

Likewise, do not represent required nested values as pointers — redesign the type:

```go
type Project struct {
	Config ProjectConfig
}

type ProjectConfig struct {
	Domain string
}
```

## Generics only for real recurring problems

Be suspicious of `Ptr[T]`, `ValueOrDefault[T]`, `ConvertSlice[T, R]`, `SafeCast[T]`, `GetOrDefault[K, V]`. Do not create generic utilities because two lines look similar. Prefer concrete domain code.

## Wrapper structs and custom primitives

Do not wrap primitives in structs without a concrete benefit:

```go
type ProjectID struct { Value string } // bad
type ProjectID string                  // fine, if it prevents mixing IDs
```

Custom string types are useful (`ProjectID`, `Region`) when they prevent accidental mixing — not for every field (`ProjectName`, `ProjectDescription`, `ProjectImage`, ...).

Enums are useful for closed sets:

```go
type DeploymentStatus string

const (
	DeploymentPending DeploymentStatus = "pending"
	DeploymentRunning DeploymentStatus = "running"
)
```

Do not enum arbitrary strings with no closed set of valid values.

## No DTO/model duplication, no mapper slop

Be suspicious when the same data has `ProjectRequest`, `ProjectDTO`, `ProjectInput`, `ProjectParams`, `ProjectData`, `ProjectModel`, `ProjectEntity`, `ProjectResponse` with nearly identical fields.

Separate types only when the contracts are materially different. Do not maintain fleets of `toDTO` / `fromDTO` / `toModel` / `fromModel` / `toEntity` / `fromEntity` when the representations are identical.

## Fix the root type instead of adding helpers

Whenever you see `asString(value)`, `asObjectID(value)`, `extractDomain(value)`, `safeValue(value)`, `toMap(value)` — trace where `value` came from and ask:

1. Why isn't this value already strongly typed?
2. Is this data external?
3. Can it be decoded into a concrete struct at the boundary?
4. Can downstream functions accept the concrete type?
5. Can this helper then disappear?

Always prefer fixing the source of poor typing.
