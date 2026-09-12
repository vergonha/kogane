---
name: boundaries
description: Where to decode and validate untrusted input — and where not to.
applies_to: "**/*.go"
---

# Boundaries

> Defend against external uncertainty, not against your own correctly typed code.

## Decode JSON once at the boundary

Bad:

```go
var payload map[string]any

if err := json.Unmarshal(body, &payload); err != nil {
	return err
}

event, _ := payload["event"].(string)
data, _ := payload["data"].(map[string]any)
projectID, _ := data["project_id"].(string)
```

Prefer:

```go
type QueueEvent struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

type ProjectSyncEvent struct {
	ProjectID string `json:"project_id"`
}

var event ProjectSyncEvent

if err := json.Unmarshal(payload.Data, &event); err != nil {
	return fmt.Errorf("decode project sync event: %w", err)
}
```

After decoding, business logic operates on typed structs.

## Remove fake defensive type assertions

```go
value, ok := data.(map[string]any)
if !ok {
	return nil
}

domain, ok := value["domain"].(string)
if !ok {
	return nil
}
```

If the data came from a contract the application controls, type it correctly upstream. If it is genuinely external, validate it once at the edge. Do not repeatedly rediscover types inside trusted application code.

## Validate where data enters — and only there

Validation is appropriate for:

- HTTP requests
- query/path parameters
- queue messages
- webhooks
- config/env vars
- external APIs
- user input
- decoded untrusted JSON
- persisted schemaless documents

Do not write validators for trusted internal structs:

```go
func validateProject(project Project) error {
	if project.ID == "" {
		return errors.New("missing project ID")
	}
	...
}
```

called from every internal service. If `Project` comes from external input, validate when creating/parsing it — once.

Never blindly remove boundary validation, security checks, or auth checks during cleanup.

## Do not reach for a validation framework

If plain Go is enough:

```go
if req.Domain == "" {
	return ErrDomainRequired
}
```

do not introduce a validation library to avoid three `if` statements. Use one only if the project already relies on it and the schema complexity warrants it.

## Keep business rules explicit

```go
// bad — the rule is hidden
if validator.IsValid(project) {

// good
if project.Status != StatusActive {
	return ErrProjectInactive
}
```

Domain rules should be visible in the code.
