---
name: safety
description: Operational limits on what a cleanup/refactor may touch.
applies_to: "**"
---

# Operational Safety

Only modify files that belong to this repository and are relevant to the cleanup.

Do not:

- modify files outside the repository
- modify secrets, credentials, `.env` files, or deployment credentials
- modify CI/CD configuration unless required for the cleanup
- modify infrastructure configuration
- modify database schemas or migrations
- modify generated code manually
- modify vendored dependencies
- upgrade or add dependencies unless explicitly required
- change build/deployment configuration
- change Docker/Kubernetes/Terraform configuration
- change public API contracts unless explicitly required by this cleanup

Do not change behavior merely to satisfy a style preference.
