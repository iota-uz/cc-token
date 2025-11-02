---
description: "Create a pull request for the current branch"
---

# Create Pull Request

Creates a pull request for the current branch with multilingual description.

## Prerequisites

- Commits must already be created on the current branch
- Current branch should be different from the base branch (usually main)

---

## Workflow

1. **Get Current Context**
    - Current branch: !`git branch --show-current`
    - Git status: !`git status`
    - Check if there are unpushed commits

2. **Ask for Base Branch**
    - Use AskUserQuestion to confirm base branch:
   ```
   AskUserQuestion:
   - Question: "Which branch should this PR target?"
   - Header: "Base Branch"
   - Options:
     - "main - Main production branch"
     - "develop - Development branch"
     - "staging - Staging branch"
   - multiSelect: false
   ```

3. **Push Branch (if needed)**
    - Check if current branch exists on remote: `git ls-remote --heads origin <current-branch>`
    - If not exists or behind: Push with `-u` flag: `git push -u origin <current-branch>`
    - If branch exists and up to date: Skip push

4. **Check Existing PR**
    - Use `mcp__github__list_pull_requests` to check if PR already exists for this branch
    - If PR exists:
        - Get PR details using `mcp__github__get_pull_request`
        - Inform user: "PR already exists: <PR_URL>"
        - Ask if they want to update the PR description (requires manual edit via GitHub UI)
        - Exit

5. **Get Commit History**
    - Get all commits from base branch to HEAD: `git log <base-branch>..HEAD --oneline`
    - Get detailed commit messages: `git log <base-branch>..HEAD --format="%H|%s|%b"`
    - Get file changes: `git diff <base-branch>...HEAD --stat`

6. **Analyze Changes**
    - Review all commits and changed files
    - Identify:
        - Main features/fixes added
        - Affected modules/domains
        - Regression areas (what could break)
        - Edge cases to test
        - User flows affected
        - Integration points touched
        - Performance/security considerations

7. **Create PR Description**
    - Generate multilingual description using template below
    - Include both English and Russian versions
    - Base content on actual commits and changes

8. **Create Pull Request**
    - Use `mcp__github__create_pull_request` with:
        - title: Concise summary from commits
        - body: Multilingual description (use HEREDOC format)
        - base: Base branch from step 2
        - head: Current branch
    - Return PR URL to user

---

## PR Description Template

Both English and Russian versions required. Analyze all commits and changes to fill this template accurately:

```markdown
## English Version

### Summary

- [First major change description based on commits]
- [Second major change description]
- [Third major change description]

### Test Plan

**Regression areas that could break:**

- [Based on analysis: what existing functionality might be affected]
- [Consider: modules touched, shared services modified, database changes]

**Edge cases to verify:**

- [Based on code changes: boundary conditions, null/empty values, etc.]
- [Consider: validation logic, error handling, async operations]

**User flows to test:**

1. [Based on UI/controller changes: step-by-step user actions]
2. [Consider: happy path + alternative paths]

**Integration points to verify:**

- [Based on dependencies: external services, APIs, database queries]
- [Consider: GraphQL resolvers, repository calls, service interactions]

**Performance/Security considerations:**

- [Based on changes: query optimization, data loading, auth checks]
- [Consider: N+1 queries, bulk operations, permission boundaries]

---

## Russian Version

### Краткое описание

- [Описание первого важного изменения на основе коммитов]
- [Описание второго изменения]
- [Описание третьего изменения]

### План тестирования

**Области регрессии, которые могут сломаться:**

- [На основе анализа: какая существующая функциональность может быть затронута]
- [Учесть: затронутые модули, измененные общие сервисы, изменения БД]

**Граничные случаи для проверки:**

- [На основе изменений кода: граничные условия, null/пустые значения и т.д.]
- [Учесть: логику валидации, обработку ошибок, асинхронные операции]

**Пользовательские сценарии для тестирования:**

1. [На основе изменений UI/контроллеров: пошаговые действия пользователя]
2. [Учесть: основной путь + альтернативные пути]

**Точки интеграции для проверки:**

- [На основе зависимостей: внешние сервисы, API, запросы к БД]
- [Учесть: GraphQL резолверы, вызовы репозиториев, взаимодействие сервисов]

**Производительность/Безопасность:**

- [На основе изменений: оптимизация запросов, загрузка данных, проверки авторизации]
- [Учесть: N+1 запросы, массовые операции, границы разрешений]

Resolves #<issue-number>
```

---

## Guidelines

### Commit Analysis

- Read ALL commits in the range (not just the latest)
- Understand the scope and purpose of changes
- Identify patterns: fix/feat/refactor/perf/etc.

### Description Quality

- **Specific**: Reference actual modules/files changed
- **Actionable**: Test plan should be clear and executable
- **Comprehensive**: Cover all changed areas
- **Realistic**: Focus on actual risks based on changes

### Example Analysis

If commits show:

- Changes to `insurance/services/osago_service.go`
- New migration `migrations/20250130_add_verification_table.sql`
- Updates to `crm/presentation/controllers/policy_controller.go`

Then description should mention:

- Summary: OSAGO verification feature, CRM policy display updates
- Regression: OSAGO quotation flow, policy list rendering
- Edge cases: Missing verification data, concurrent policy updates
- User flows: OSAGO quotation → verification → policy creation
- Integration: Verification service ↔ OSAGO service ↔ Policy repository
- Performance: New table indexes, query optimization in policy list

---

## Error Handling

**No commits found:**

- Error: "No commits on current branch compared to base branch"
- Suggest: Check if you're on the correct branch or need to create commits first

**PR already exists:**

- Show existing PR URL
- Exit (no duplicate PR creation)

**Push fails:**

- Check network connection
- Check if branch protection rules block push
- Suggest: Pull the latest changes and retry

**GitHub API errors:**

- Check authentication: `gh auth status`
- Check repository permissions
- Provide clear error message to user

---

