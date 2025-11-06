# Git Workflow Summary

## What's Been Set Up

Your repository now has comprehensive git commit management rules that integrate seamlessly with Cursor 2.0's Agent capabilities.

### 📁 Files Created/Updated

1. **`.cursor/rules/git_commit_management.mdc`**
   - Comprehensive rule for Cursor Agent
   - Always applied to all files
   - Enforces atomic commits and Conventional Commits

2. **`GIT_COMMIT_CHEATSHEET.md`**
   - Quick reference guide at repo root
   - Easy to access when writing commits
   - Common patterns and examples

3. **`README.md`** (Updated)
   - Added git workflow to Contributing section
   - Links to documentation

4. **`Makefile`** (Updated)
   - New git workflow helpers
   - Pre-commit validation commands

## 🚀 Quick Start

### Daily Workflow with Cursor Agent

```bash
# NEW: Agent asks permission before committing

# 1. Tell agent what to do
You: "Add product variants feature"

# 2. Agent implements and asks permission
Agent: "Ready to commit feat(product): add variant model. Proceed?"

# 3. You approve
You: "yes"

# 4. Agent commits automatically
Agent: "✓ Committed"

# 5. Agent continues to next layer and asks again
Agent: "Ready to commit feat(product): add variant repository. Proceed?"
You: "yes"

# 6. Review your commits anytime
make git-log

# ALTERNATIVE: Manual commits if you prefer
You: "Implement variants but don't commit"
Agent: *implements*
You: *manually review and commit*
```

### Working with the Agent

**Tell the Agent:**
- ✅ "Implement product variants" (agent will ask permission to commit each layer)
- ✅ "Add tests for variants but don't commit yet" (manual control)
- ✅ "Refactor category service and commit when ready"

**Agent Will:**
- ✅ Focus on one domain/layer at a time
- ✅ Ask permission before committing: "Ready to commit X. Proceed?"
- ✅ Show files and commit message before asking
- ✅ Execute commit after you say "yes"
- ❌ Never commit without asking first

## 📋 New Makefile Commands

```bash
# View status with domain grouping
make git-status

# View recent commit history with graph
make git-log

# View staged vs unstaged changes
make git-diff

# Pre-commit validation (tests + lint)
make git-check
```

## 🎯 Key Principles

### 1. Atomic Commits by Domain

**DO**: Separate commits per domain
```bash
git commit -m "feat(product): add variant model"
git commit -m "feat(leaflet): integrate variants"
git commit -m "feat(api): add variant endpoints"
```

**DON'T**: One big commit
```bash
# BAD - Hard to review and revert
git add .
git commit -m "add variant feature"
```

### 2. Conventional Commits Format

```
<type>[scope]: <description>

[optional body]

[optional footer]
```

**Types**: `feat`, `fix`, `refactor`, `test`, `docs`, `perf`, `build`, `chore`
**Scopes**: Domain or layer names (`product`, `api`, `db`, etc.)

### 3. Commit Granularity

**Ideal**: < 10 files, < 500 lines per commit
**Maximum**: ONE domain + tests + docs

## 📖 Examples

### Feature Across Multiple Domains

```bash
# Scenario: Add pricing feature affecting promotions and products

# Commit 1: Core domain (promotion)
git add internal/domain/promotion/model/
git commit -m "feat(promotion): add dynamic pricing model"

# Commit 2: Integration domain (product)
git add internal/domain/product/service.go
git commit -m "feat(product): integrate dynamic pricing"

# Commit 3: API layer
git add internal/api/handler/promotion/pricing.go
git commit -m "feat(api): add pricing calculation endpoint"

# Commit 4: Tests
git add internal/domain/promotion/*_test.go
git commit -m "test(promotion): add pricing calculation tests"

# Commit 5: Documentation
git add docs/PRICING.md
git commit -m "docs(promotion): document pricing feature"
```

### Bug Fix

```bash
# Single commit is OK for focused bug fix
git add internal/domain/product/service.go internal/api/handler/product/search.go
git commit -m "fix(product): prevent nil pointer in search results

Added null checks for optional product fields and proper
error handling when product data is incomplete.

Fixes #789"
```

### Database Migration

```bash
git add db/migrations/0004_add_variants.sql
git commit -m "build(db): add migration for product variants"

git add internal/domain/product/model/
git commit -m "feat(product): update model for variants"

git add internal/domain/product/repository/
git commit -m "feat(product): update repository for variants"
```

## 🔍 Pre-Commit Checklist

Before every commit, verify:

- [ ] **Scope**: One domain/layer/concern only?
- [ ] **Tests**: Do all tests pass? (`make test`)
- [ ] **Lints**: No linting issues? (`make lint`)
- [ ] **Message**: Follows Conventional Commits?
- [ ] **Size**: < 10 files, < 500 lines?
- [ ] **Revert**: Can be safely reverted independently?

**Quick validation:**
```bash
make git-check
```

## 🎨 Good vs Bad Examples

### ✅ Good Commit History

```
feat(product): add variant model
feat(product): add variant repository
feat(product): add variant service
feat(api): add variant endpoints
test(product): add variant tests
docs(product): document variant feature
```

**Why Good:**
- Each commit has clear scope
- Logical progression
- Easy to review and revert
- Tells a story

### ❌ Bad Commit History

```
wip
more stuff
fix
final changes
updates
```

**Why Bad:**
- Vague, uninformative
- Hard to understand what changed
- No context for future developers
- Breaks conventional commits

## 🔧 Fixing Mistakes

### Forgot to Split Commits

```bash
# Undo last commit, keep changes staged
git reset --soft HEAD~1

# Now stage and commit in smaller pieces
git add internal/domain/product/
git commit -m "feat(product): add variant model"

git add internal/domain/leaflet/
git commit -m "feat(leaflet): integrate variants"
```

### Wrong Commit Message

```bash
# Amend the last commit message
git commit --amend
```

### Forgot Files in Commit

```bash
# Add forgotten files
git add forgotten_file.go

# Amend without changing message
git commit --amend --no-edit
```

### Squash Multiple WIP Commits

```bash
# Interactive rebase for last 5 commits
git rebase -i HEAD~5

# In editor, change 'pick' to 'squash' for commits to merge
# Then write a proper conventional commit message
```

## 🤖 Integration with Cursor Agent

### Agent Behavior (Configured)

The agent is now configured to:

1. **Focus on one domain at a time** when implementing features
2. **Notify you** when a logical commit point is reached
3. **Suggest commit messages** following Conventional Commits
4. **Ask permission before committing** - shows what will be committed and waits for approval
5. **Execute commits after approval** - runs git add + git commit after you say "yes"
6. **Keep changes atomic** and reviewable

### Instructing the Agent

**Recommended Flow (With Auto-Commit):**
```
You: "Implement product variants feature"

Agent: *implements product model*
Agent: "Changes ready. Ready to commit as:
        'feat(product): add variant model'

        Files to be committed:
        - internal/domain/product/model/variant.go
        - internal/domain/product/model/model.go

        Proceed? (yes/no)"

You: "yes"

Agent: *commits automatically*
Agent: "✓ Committed: feat(product): add variant model"
Agent: *continues with repository layer*
```

**Alternative (Manual Commit):**
```
You: "Implement product variants but don't commit yet"

Agent: *implements all changes*
Agent: "Changes ready. Suggested commits:
        1. feat(product): add variant model
        2. feat(product): add variant repository
        3. feat(api): add variant endpoints"

You: *manually review and commit each*
```

**For Large Features:**
```
You: "Break down the variant feature into commits and ask before each"

Agent: *implements model layer*
Agent: "Ready to commit feat(product): add variant model. Proceed?"
You: "yes"
Agent: *commits*

Agent: *implements repository layer*
Agent: "Ready to commit feat(product): add variant repository. Proceed?"
You: "yes"
Agent: *commits*
```

## 📚 Additional Resources

- **Full Documentation**: `.cursor/rules/git_commit_management.mdc`
- **Quick Reference**: `GIT_COMMIT_CHEATSHEET.md`
- **Conventional Commits**: https://www.conventionalcommits.org/
- **Contributing Guide**: See `README.md` Contributing section

## 💡 Pro Tips

1. **Use `git add -p`** for interactive staging (pick specific changes)
2. **Review staged changes** before committing: `git diff --cached`
3. **Keep feature branches clean** with `git rebase -i` before PR
4. **Write commit messages for future you** - explain "why", not just "what"
5. **Link to issues** in commit footers: `Fixes #123`
6. **Mark breaking changes** with `!` or `BREAKING CHANGE:` footer
7. **Use descriptive scopes** - `product` not `prod`, `api` not `a`
8. **Commit often** - it's easier to squash than to split

## 🎯 Success Metrics

You'll know this workflow is working when:

- ✅ Each commit can be understood in isolation
- ✅ Git history tells the story of your project
- ✅ Reverting changes is safe and surgical
- ✅ Code reviews are faster (reviewers can review commit-by-commit)
- ✅ You can generate changelogs automatically
- ✅ You can trace features to commits to issues

## 🚨 Common Pitfalls to Avoid

1. **The "Everything" Commit**: `git add .` across multiple domains
2. **The "And" Commit**: Message contains "and" (sign of multiple concerns)
3. **The "WIP" Habit**: Leaving WIP commits in main/feature branch
4. **The "Vague" Message**: "fix bug" or "update code"
5. **The "Cross-Domain"**: Changes in product + leaflet + promotion in one commit
6. **The "No Scope"**: `feat: add feature` (missing scope)
7. **The "Past Tense"**: "added feature" instead of "add feature"

## 📞 Getting Help

If you're unsure about:
- How to split a large change: See examples in `.cursor/rules/git_commit_management.mdc`
- Commit message format: Check `GIT_COMMIT_CHEATSHEET.md`
- What to commit together: Look at the "Commit Granularity" section

---

**Remember**: Good commits are a gift to your future self and your team. The extra 30 seconds to craft a good commit message pays dividends for months to come.

Happy committing! 🎉

