# Git Commit Cheatsheet

Quick reference for Conventional Commits.

## Format

```
<type>[scope]: <description>
```

## Types & Scopes

| Type | Scope Examples | Example Message |
|------|----------------|-----------------|
| `feat` | `product`, `api`, `leaflet` | `feat(product): add variant model` |
| `fix` | `product`, `category` | `fix(leaflet): prevent duplicate associations` |
| `refactor` | `category`, `service` | `refactor(category): extract tree builder` |
| `test` | `promotion`, `product` | `test(promotion): add validation tests` |
| `docs` | `api`, `product` | `docs(api): document auth flow` |
| `build` | `db`, `deps`, `docker` | `build(db): add migration for variants` |

**Domains**: `product`, `leaflet`, `promotion`, `category`, `distributor`
**Layers**: `api`, `service`, `repository`, `middleware`
**Infra**: `db`, `migration`, `config`, `docker`

## Rules

✅ **DO**: One domain/layer per commit, imperative mood, < 72 chars, lowercase, no period
❌ **DON'T**: Multiple domains, "WIP", capitalize, use "and"

## Examples

### Feature Across Layers
```bash
git commit -m "feat(product): add variant model"
git commit -m "feat(product): add variant repository"
git commit -m "feat(api): add variant endpoints"
git commit -m "test(product): add variant tests"
```

### Bug Fix
```bash
git commit -m "fix(product): prevent nil pointer in search"
```

### Database Change
```bash
git commit -m "build(db): add migration for product variants"
git commit -m "feat(product): update model for variants"
```

## Quick Commands

```bash
make git-status    # Status with domain grouping
make git-check     # Pre-commit validation
git diff --cached  # Review staged changes
git add -p file.go # Stage interactively
```

---

**Full docs**: `.cursor/rules/git_commit_management.mdc`

