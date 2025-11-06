# Release Please Integration Guide

## What is Release Please?

Release Please is an automated release management tool from Google that:
- **Automates version bumping** based on Conventional Commits
- **Generates CHANGELOG.md** automatically from commit messages
- **Creates GitHub Releases** with detailed release notes
- **Handles semantic versioning** (MAJOR.MINOR.PATCH)

## How It Works

### 1. Commit to Main Branch

When you merge commits to `main` using Conventional Commits format:

```bash
feat(publisher): add batch publishing support
fix(subscriber): prevent race condition in consumer groups
refactor(redis): extract connection pool logic
```

### 2. Release Please Creates a PR

Release Please automatically:
1. **Analyzes commits** since the last release
2. **Calculates version bump** based on commit types:
   - `feat` → Minor version bump (0.1.0 → 0.2.0)
   - `fix` → Patch version bump (0.1.0 → 0.1.1)
   - `feat!` or `BREAKING CHANGE:` → Major version bump (0.1.0 → 1.0.0)
3. **Creates a Release PR** with:
   - Updated version in `.release-please-manifest.json`
   - Generated CHANGELOG.md entries
   - All commits grouped by type

### 3. You Merge the Release PR

When you merge the Release PR:
1. **GitHub Release is created** automatically
2. **Git tag** is created (e.g., `v0.2.0`)
3. **Release notes** are published with all changes

### 4. Continuous Cycle

The cycle repeats with every merge to `main`:
```
Commits → Main → Release PR → Merge → GitHub Release → Repeat
```

## Difference from Manual Workflow

### Manual Workflow (What We Built)
- **Focus**: Local development and PR creation
- **Agent commits** iteratively during development
- **You review and approve** each commit
- **You write** PR descriptions manually (or with agent help)
- **Enforces**: Atomic commits, Conventional Commits format
- **Output**: Clean commit history, reviewable PRs

### Release Please (Automated Releases)
- **Focus**: Release automation after merging to main
- **Reads commit history** from merged PRs
- **Automatically generates** changelog and release notes
- **Automatically bumps** version based on commit types
- **Creates**: Release PRs, GitHub Releases, Git tags
- **Output**: Semantic versioning, changelogs, releases

## How They Work Together

```
┌─────────────────────────────────────────────────────────────┐
│                    Development Phase                         │
│          (Your Git Workflow + Cursor Agent)                  │
└─────────────────────────────────────────────────────────────┘
                            ↓
    ┌───────────────────────────────────────────┐
    │  Agent implements feature in atomic commits│
    │  - feat(publisher): add batch publishing   │
    │  - feat(redis): implement batch publishing │
    │  - test(publisher): add batch tests        │
    │  - docs(examples): add batch example       │
    └───────────────────────────────────────────┘
                            ↓
    ┌───────────────────────────────────────────┐
    │  You push branch, create PR               │
    │  Agent suggests PR description            │
    │  - Summary of changes                     │
    │  - Grouped by component                   │
    │  - Testing instructions                   │
    └───────────────────────────────────────────┘
                            ↓
    ┌───────────────────────────────────────────┐
    │  Code Review & Merge to Main              │
    │  All commits preserved in history         │
    └───────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│                    Release Phase                             │
│                  (Release Please)                            │
└─────────────────────────────────────────────────────────────┘
                            ↓
    ┌───────────────────────────────────────────┐
    │  Release Please analyzes commits          │
    │  - Found 1 feat commit → minor bump       │
    │  - Current: v0.1.0 → Next: v0.2.0        │
    └───────────────────────────────────────────┘
                            ↓
    ┌───────────────────────────────────────────┐
    │  Release Please creates Release PR        │
    │  - Updates .release-please-manifest.json  │
    │  - Generates CHANGELOG.md entries         │
    │  - Groups changes by type                 │
    └───────────────────────────────────────────┘
                            ↓
    ┌───────────────────────────────────────────┐
    │  You merge Release PR                     │
    │  GitHub Release v0.2.0 is created         │
    │  Git tag v0.2.0 is created                │
    └───────────────────────────────────────────┘
```

## Key Differences

| Aspect | Your Git Workflow | Release Please |
|--------|-------------------|----------------|
| **When** | During development | After merge to main |
| **Who** | You + Cursor Agent | GitHub Actions (automated) |
| **Input** | Code changes | Commit history |
| **Output** | Commits + PRs | Releases + Tags + Changelog |
| **Purpose** | Clean, reviewable commits | Automated version management |
| **Manual Steps** | Approve commits, write PR descriptions | Review and merge Release PRs |

## Configuration Files

### `.github/workflows/release-please.yml`
GitHub Actions workflow that runs Release Please on every push to `main`.

### `release-please-config.json`
Configuration for Release Please:
- **Release type**: `go` (for Go projects)
- **Changelog sections**: Maps commit types to changelog sections
- **Version bumping rules**: How to calculate semantic versions

### `.release-please-manifest.json`
Tracks current version. Release Please updates this file.

## Example Release Cycle

### Step 1: Merge Feature PR
```bash
# Your PR with 4 commits merged to main:
feat(publisher): add batch publishing interface
feat(redis): implement batch publishing for Redis
test(publisher): add batch publishing tests
docs(examples): add batch publishing example
```

### Step 2: Release Please Creates PR
**Title**: `chore(main): release 0.2.0`

**CHANGELOG.md** excerpt:
```markdown
## [0.2.0](https://github.com/you/promy-event-bus/compare/v0.1.0...v0.2.0) (2025-11-06)

### Features

* **publisher**: add batch publishing interface ([abc1234](link))
* **redis**: implement batch publishing for Redis ([def5678](link))

### Tests

* **publisher**: add batch publishing tests ([ghi9012](link))

### Documentation

* **examples**: add batch publishing example ([jkl3456](link))
```

### Step 3: You Merge Release PR

**GitHub Release v0.2.0** is created with:
- Release notes (same as CHANGELOG)
- Git tag `v0.2.0`
- Downloadable assets (if configured)

## Benefits

1. **No Manual Version Bumping**: Automatic based on commits
2. **Auto-Generated Changelog**: Always up-to-date
3. **Consistent Release Process**: Same workflow every time
4. **Semantic Versioning**: Automatic MAJOR.MINOR.PATCH
5. **Audit Trail**: Clear history of what changed when
6. **Integration Ready**: Works with CI/CD, deployment automation

## When to Merge Release PRs

**Recommended Strategy:**
- **Weekly releases**: Merge Release PR every Friday
- **On-demand**: Merge when critical fixes are ready
- **Before deployments**: Merge before deploying to production

**Note**: You don't *have* to merge Release PRs immediately. Release Please keeps updating the same PR until you merge it.

## Breaking Changes

For breaking changes, use `!` or `BREAKING CHANGE:` footer:

```bash
# Option 1: ! after type
feat(publisher)!: change publish method signature

# Option 2: BREAKING CHANGE footer
feat(publisher): change publish method signature

BREAKING CHANGE: Publish now requires context as first parameter.
Clients must update to pass context.Context.
```

This triggers a **major version bump** (0.2.0 → 1.0.0).

## Summary

- **Your Git Workflow**: Ensures clean, atomic commits during development
- **Release Please**: Automates releases after merging to main
- **Together**: Complete automation from development to release
- **You Control**: When to merge Release PRs and deploy

The combination gives you:
1. **Clean commit history** (from your workflow)
2. **Automated releases** (from Release Please)
3. **Professional changelogs** (from both)
4. **Semantic versioning** (automatic)
5. **Minimal manual work** (maximum automation)

---

**Next Steps**:
1. Push these changes to GitHub
2. Merge a PR to `main` with Conventional Commits
3. Watch Release Please create your first Release PR
4. Review and merge the Release PR
5. See your first automated GitHub Release!

