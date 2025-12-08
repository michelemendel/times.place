# times.place

A web application for listing venues and their event times, built using **Spec-Driven Development (SDD)** methodology.

## Project Overview

**times.place** is a venue management and event scheduling platform designed to serve communities by providing accurate event times (e.g. prayer times).

This is a prototype web application that allows:

- Visitors to browse venues and view upcoming event times, contact details, and venue information
- Admins to edit venue details, schedules, and content through a simple admin interface
- Bilingual support for English and Hebrew content

The current prototype is frontend-only, using Svelete and SvelteKit with local storage for data persistence. Future versions will include a Go backend with PostgreSQL.

## Directory Structure

A high-level overview of the codebase:

```bash
times.place/
├── blueprint/       # SDD documentation (Single Source of Truth)
├── cmd/             # Go application entry points (planned)
├── docs/            # General project documentation
├── domain/          # Domain logic and entities
├── minyanim/        # Examples of how synagogues are presenting the prayer times. These are the documents that triggered the idea of times.place.
├── frontend/        # SvelteKit frontend source code
├── static/          # Static assets
└── utils/           # Backend (Go) utility functions (planned)
```

## Spec-Driven Development (SDD) Methodology

The project follows **Spec-Driven Development**, a methodology where documentation drives the development process. Instead of (vibe-)coding first and documenting later, we:

1. **Govern** the process through `constitution.md` (governance, principles, and process rules goes here)
2. **Specify** requirements, user stories, and success criteria in `spec.md`
3. **Plan** technical approaches and implementation decisions in `plan.md`
4. **Track** tasks and progress in `tasks.md`
5. **Implement** following the spec and plan, logging decisions in `implement.md`

The `blueprint/` folder contains the core SDD documentation that guides all development. All development should reference and update these files as work progresses.

**This approach ensures:**

- Clear requirements before coding begins
- Documented technical decisions and rationale
- Better collaboration between human developers and AI agents
- Maintainable code with clear reasoning

### SDD Resources

**General SDD Resources:**

- [Spec-driven development with AI: Get started with a new open source toolkit](https://github.blog/ai-and-ml/generative-ai/spec-driven-development-with-ai-get-started-with-a-new-open-source-toolkit/) - GitHub's introduction to SDD
- [A Practical Guide to Spec-Driven Development](https://docs.zencoder.ai/user-guides/tutorials/spec-driven-development-guide) - Comprehensive guide from Zencoder
- [spec-kit](https://github.com/github/spec-kit) - GitHub's open-source toolkit for SDD

## Tech Stack

- **Frontend**: SvelteKit
- **Styling**: Tailwind CSS
- **Data Storage**: LocalStorage (prototype), PostgreSQL (planned)
- **Backend**: Go (planned)
- **Hosting**: Render.com

## My Notes

- [Perplexity Space: times.place](https://www.perplexity.ai/spaces/times-place-0QMkL8qGR16UjuPLu0HfIw)
- [Svelte Notes](docs/svelte.md) - My notes while learning Svelte

## Getting Started

(Development setup instructions will be added as the project progresses)
