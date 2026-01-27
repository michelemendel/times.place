---
name: update-backend-implement-log
description: Updates blueprint/backend/4_implement.md with latest backend implementation work. Use when backend code changes have been made and need to be documented, when the user asks to update the implementation log, or after completing backend implementation tasks.
---

# Update Backend Implementation Log

## Purpose

Updates `blueprint/backend/4_implement.md` with a new dated entry documenting recent backend implementation work, following the established format and template.

## Workflow

### 1. Read Current Implementation Log

**CRITICAL**: First, read the entire `blueprint/backend/4_implement.md` file to understand:
- What work has already been documented
- The date of the last entry
- The format and style of existing entries

This prevents documenting work that's already been logged.

### 2. Identify Recent Changes

Discover what backend work has been completed since the last entry:

- **Check git history**: `git log --since="<last-entry-date>" --oneline -- backend/` to see commits
- **Review file changes**: Check for new/modified files in `backend/` directory
- **Check key areas**:
  - `backend/db/migrations/` - new migration files
  - `backend/db/queries/` - new SQL queries
  - `backend/db/sqlc/` - generated code changes
  - `backend/cmd/` - new CLI tools or API handlers
  - `backend/internal/` - service layer, auth, test helpers
  - `Makefile` - new database or build targets
  - `backend/README.md` or other docs - documentation updates

**IMPORTANT**: Compare git commits and file changes against what's already documented in the log. Only document NEW work that hasn't been logged yet. If a commit's work is already documented, skip it.

### 3. Determine Today's Date

Use today's date in `YYYY-MM-DD` format for the new entry header.

### 4. Organize Changes by Category

Group changes into the standard note categories:

- **Migrations**: New migration files, schema changes, rollback fixes
- **sqlc**: New query files, generated code updates, query changes
- **Auth/JWT**: Authentication implementation, token handling, middleware
- **API**: New endpoints, handler changes, request/response updates
- **Dev container**: Docker changes, devcontainer config, port mappings
- **Dev workflow**: Local development setup, HMR, build process
- **Production**: Deployment config, environment variables, Render.com setup
- **Test data**: Seeding infrastructure, test helpers, fixtures
- **Makefile**: New targets, command changes

### 5. Write Summary

Create a concise summary answering:
- **What changed?** - High-level overview of work completed
- **Why?** - Reason or motivation (if relevant)

Use bullet points with bold keywords for key areas.

**IMPORTANT**: Only include work that is NOT already documented in the log. If you find that all recent work is already documented, inform the user that there's nothing new to add.

### 6. Write Notes Section

For each relevant category, provide specific details:

- List files created/modified
- Describe functionality added
- Note any deviations from `backend/2_plan.md`
- Include technical details that would be useful for future reference

Use nested bullets for sub-items within categories.

### 7. Add Entry to File

**CRITICAL**: Insert the new entry at the END of the file (after the last dated entry), NOT at the top. Entries are in chronological order with the most recent at the bottom.

Follow this format:

```markdown
## YYYY-MM-DD

### Summary

- [What changed?]
- [Why?]

### Notes

- **Category**: Details...
  - Sub-item if needed
- **Category**: Details...
```

## Example

```markdown
## 2026-01-27

### Summary

- **Database schema & migrations**: Created complete database schema with 4 migrations covering all core tables, indexes, constraints, and triggers.
- **Makefile database commands**: Set up comprehensive Makefile targets for database operations (migrations, seeding, verification, connection).

### Notes

- **Migrations** (created 4 migration files in `backend/db/migrations/`):
  - `00001_enable_pgcrypto.sql`: Enables pgcrypto extension for UUID generation
  - `00002_create_tables.sql`: Creates core schema...
- **Makefile database targets**:
  - `dbgooseup`: Apply all pending migrations
  - `dbgoosedown`: Rollback last migration (one at a time)
```

## Important Notes

- **CRITICAL**: Always read the entire log file first to see what's already documented
- **CRITICAL**: Only document NEW work that hasn't been logged yet - compare git commits against existing entries
- Always use today's date (YYYY-MM-DD format)
- **Place new entries at the END of the file** (after the last dated entry), NOT at the top - entries are in chronological order
- Follow the existing format and style of previous entries
- Be specific about file paths and technical details
- Include context about why changes were made if it deviates from the plan
- If multiple distinct work sessions happened on the same day, use separate `### Summary` sections under the same date header
- If all recent work is already documented, inform the user rather than creating a duplicate entry
