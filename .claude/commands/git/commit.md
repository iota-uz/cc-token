---
allowed-tools: |
  Bash(git status:*), Bash(git diff:*), Bash(git log:*),
  Bash(git add:*), Bash(git commit:*), Bash(git push:*), Bash(git pull:*),
  Bash(git checkout:*), Bash(git branch:*), Bash(make check:*), Bash(templ generate:*),
  Bash(gh pr:*),
  Read, Edit, AskUserQuestion
description: "Commit changes with option to push to current branch or create new branch with PR"
---

## Workflow Sequence

1. Ask user about branching strategy
2. Check if CHANGELOG.md should be updated
3. Pre-Commit Preparation
4. Create commits
5. Post-Commit Actions

## Current codebase status:

- Current git status: !`git status --porcelain`
- Changed files: !`git diff --name-only`
- Current git diff: !`git diff`
- Current git branch: !`git branch --show-current`
- Recent commits: !`git log --oneline -5`

## Branch Strategy

Always ask the user about branch strategy using the AskUserQuestion tool:

```
AskUserQuestion:
- Question: "How would you like to proceed?"
- Header: "Branch Strategy"
- Options:
  - "Push to current branch - Commit and push to current branch"
  - "Create new branch + PR - Create new feature branch first, then commit and push"
- multiSelect: false
```

**If "Push to current branch":**

- Continue directly to "Context Loading & CHANGELOG Check"
- After commits created, push to current branch (see Post-Commit Actions)

**If "Create new branch + PR":**

- Ask user for new branch name with suggestion based on changed files
    - Example suggestions: `fix/driver-form-validation`, `feat/load-tracking`, `refactor/finance-services`
- Create and checkout new branch BEFORE any other steps
- Then continue to "Context Loading & CHANGELOG Check"
- After commits created, push new branch with `-u` flag and create PR (see Post-Commit Actions)

## CHANGELOG Check (After Branch Created)

Evaluate if CHANGELOG.md should be updated. This happens AFTER branch creation but BEFORE pre-commit preparation:

**Count changed files by layer:**

- Presentation: `*_controller.go`, `*_viewmodel.go`, `*.templ`, `*.toml`
- Business: `*_service.go`, domain models
- Infrastructure: `*_repository.go`, `migrations/*.sql`

**MUST update if ANY:**

- 3+ files changed across 2+ layers (multi-layer feature)
- Files in `migrations/*.sql` changed (database migrations)
- Commit message includes "breaking:", "security:", or "BREAKING CHANGE:"
- New module/domain directory created

**SHOULD update if ANY:**

- 5+ files changed in single layer (significant scope)
- Commit type is `feat:` or `refactor:` with `*_service.go` changes
- Third-party integration files modified
- Performance-related changes in core business logic

**SKIP if:**

- Only test files changed (`*_test.go`)
- Only documentation changed (`.md` files)
- Single-file cosmetic/formatting changes
- Only dependency updates (go.mod, go.sum)

**If MUST criteria met:**

- Inform user CHANGELOG update is required
- Use AskUserQuestion to get change description:
  ```
  AskUserQuestion:
  - Question: "Describe this change for CHANGELOG (≤300 characters)"
  - Header: "CHANGELOG Entry"
  - Options:
    - "feat: [your description]"
    - "fix: [your description]"
    - "refactor: [your description]"
    - "perf: [your description]"
    - "security: [your description]"
  - multiSelect: false
  ```
- Update CHANGELOG.md with new entry (date | type | description)
- Enforce FIFO: Remove oldest if >10 entries
- Stage CHANGELOG.md for commit

**If SHOULD criteria met:**

- Use AskUserQuestion to ask if user wants to update:
  ```
  AskUserQuestion:
  - Question: "This looks like a significant change. Update CHANGELOG.md?"
  - Header: "CHANGELOG Update"
  - Options:
    - "Yes, update CHANGELOG"
    - "No, skip for now"
  - multiSelect: false
  ```
- If "Yes": Follow same process as MUST criteria
- If "No": Continue to Pre-Commit Preparation

---

## Pre-Commit Preparation (CRITICAL)

1. Analyze changed files to understand the nature of changes
2. If `.templ` files were changed, regenerate them using `templ generate`
    - If regeneration fails: report errors and stop
3. If `.go` or `.templ` files were changed, format them using `make fix fmt`
    - If formatting fails: report errors and stop
4. If `.toml` files were changed, test them using `make check tr`
    - If validation fails: ask the user how to proceed
5. Delete build artifacts or temporary files (ask user if unsure)
    - Never commit items from "Never Commit" section below
6. Verify all changes are legitimate (no secrets, no build artifacts)

If any step fails: Stop and report errors. Do not proceed to commits.

## Commit Messages

Use conventional commit prefixes:

- `fix:` - Bug fixes
- `feat:` - New features
- `docs:` - Documentation updates (.md files)
- `ci:` - CI/CD configuration changes (Dockerfile, .github/workflows/**/, stack.yml, compose.*.yml)
- `wip:` - Work in progress
- `style:` - Code formatting (no functional changes)
- `perf:` - Performance improvements
- `test:` - Adding or updating tests
- `chore:` - Maintenance tasks
- `refactor:` - Code restructuring without changing functionality
- `cc:` - Claude code configuration files (CLAUDE.md, .claude/**/)

**Guidelines:**

- Clear, concise messages explaining "what" and "why"
- Under 50 characters when possible
- Present tense and imperative mood

---

## Never Commit

- Build artifacts (binaries, docker images, etc.)
- Test coverage reports (coverage.out, coverage.html)
- Temporary files (*.test, *.out, *.go.old, etc.)
- Log files (logs/app.log, etc.)
- IDE files (*.swp, *.swo, *.swn, etc.)
- Root markdown files (FOLLOW_UP_ISSUES.md, PR-*-REVIEW.md)
- CSV/Images/Docs unless explicitly told to do so
- Generated files (except `*_templ.go`)

## Creating Commits

After pre-commit preparation succeeds:

1. Group related changes into logical commits
2. Create multiple commits if changes span different features or fixes
3. For each commit:
    - Use appropriate conventional prefix (fix:, feat:, refactor:, etc.)
    - Include clear, concise message
    - Stage related files: `git add [files]`
    - Commit with message: `git commit -m "prefix: message"`
4. Commit `*_templ.go` files even though they are generated
5. Include CHANGELOG.md in commit if it was updated

## Post-Commit Actions

### If "Push to Current Branch"

1. Pull the latest changes from remote (CRITICAL)
2. Push commits to the current branch
3. Report the successful push with commit count

**Error Handling:**

- Pull conflicts → stop and inform user (commits are safe, needs manual conflict resolution)
- Push fails → suggest resolving conflicts or checking network

### If "Create New Branch + PR"

1. New branch already created before pre-commit phase
2. Push new branch with `-u` flag: `git push -u origin <new-branch>`
3. Check if PR already exists for this branch: `gh pr list --head <new-branch> --json number,url`
4. If NO PR exists:
    - Get commit history: `git log <original-branch>..HEAD` (where original-branch is the branch user started from)
    - Analyze commits to create multilingual PR description (see template below)
    - Create PR using `gh pr create --title "..." --body "$(cat <<'EOF'...)"` with multilingual description
    - Return PR URL from command output
5. If PR exists:
    - Just push commits (PR auto-updates)
    - Return existing PR URL from list output

## PR Description Template

Both English and Russian versions required:

```markdown
## English Version

### Summary

- First change description
- Second change description
- Third change description

### Test Plan

**Regression areas that could break:**

- Area 1 description
- Area 2 description

**Edge cases to verify:**

- Case 1 description
- Case 2 description

**User flows to test:**

1. Step 1 description
2. Step 2 description

**Integration points to verify:**

- Integration 1 description
- Integration 2 description

**Performance/Security considerations:**

- Consideration 1 description
- Consideration 2 description

## Russian Version

### Краткое описание

- Описание первого изменения
- Описание второго изменения
- Описание третьего изменения

### План тестирования

**Области регрессии, которые могут сломаться:**

- Описание области 1
- Описание области 2

**Граничные случаи для проверки:**

- Описание случая 1
- Описание случая 2

**Пользовательские сценарии для тестирования:**

1. Описание шага 1
2. Описание шага 2

**Точки интеграции для проверки:**

- Описание интеграции 1
- Описание интеграции 2

**Производительность/Безопасность:**

- Описание соображения 1
- Описание соображения 2

Resolves #<issue-number>
```

