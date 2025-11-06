# Git Commit Cheatsheet

Quick reference for Conventional Commits.

## Format

```
<type>[scope]: <description>
```

## Types & Scopes

| Type | Scope Examples | Example Message |
|------|----------------|-----------------|
| `feat` | `publisher`, `subscriber`, `events` | `feat(publisher): add batch publishing` |
| `fix` | `publisher`, `subscriber`, `redis` | `fix(subscriber): prevent race condition` |
| `refactor` | `redis`, `events` | `refactor(redis): extract connection pool` |
| `test` | `publisher`, `redis`, `events` | `test(publisher): add batch publishing tests` |
| `docs` | `examples`, `readme` | `docs(examples): add batch publishing example` |
| `build` | `deps`, `docker`, `ci` | `build(deps): upgrade redis client to v9.0.0` |

**Core Components**: `publisher`, `subscriber`, `event`, `config`, `errors`
**Event Schemas**: `events` (for event schemas in events/ package)
**Implementations**: `redis` (for Redis implementation)
**Infrastructure**: `examples`, `testutil`, `docs`, `build`, `ci`

## Rules

✅ **DO**: One component/concern per commit, imperative mood, < 72 chars, lowercase, no period
❌ **DON'T**: Multiple components, "WIP", capitalize, use "and"

## Examples

### Feature Across Components
```bash
git commit -m "feat(publisher): add batch publishing interface"
git commit -m "feat(redis): implement batch publishing for Redis"
git commit -m "test(publisher): add batch publishing tests"
```

### Bug Fix
```bash
git commit -m "fix(subscriber): prevent race condition in consumer groups"
```

### Event Schema Change
```bash
git commit -m "feat(events): add product identified event schema"
git commit -m "test(events): add product event validation tests"
```

### Refactoring
```bash
git commit -m "refactor(redis): extract connection pool logic"
```

## Quick Commands

```bash
make git-status    # Status with component grouping
make git-check     # Pre-commit validation
git diff --cached  # Review staged changes
git add -p file.go # Stage interactively
```

---

**Full docs**: `.cursor/rules/git_commit_management.mdc`

