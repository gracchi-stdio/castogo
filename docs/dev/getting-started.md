---
title: Getting Started
description: Set up the castogo development environment.
order: 1
---

# Getting Started

Welcome to the **castogo** developer docs. This is a sample file so you can see
how markdown with frontmatter renders inside the admin before adding real content.

## Prerequisites

- Go 1.26+
- Bun (frontend toolchain)
- PostgreSQL 17
- `just` task runner

## Common Commands

| Command | Purpose |
|---------|---------|
| `just dev` | Vite HMR + templ watch + air live reload |
| `just generate` | sqlc + templ codegen |
| `just build` | production build (also the embed gate) |

## Next steps

Replace this file with real developer documentation under `docs/dev/`. Each file
needs a YAML frontmatter block with `title`, `description`, and `order`.
