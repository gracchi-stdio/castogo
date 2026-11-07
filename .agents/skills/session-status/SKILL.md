---
name: session-status
description: "Resume or compact an agent session on this repo. New thread / 'resume' / 'where were we': read STATUS.md next to this file and continue from its Next section. 'compact' or a finished step: rewrite STATUS.md (decisions, repo state, Anytype docs, next steps) — replace, never append. Cross-thread memory for the castogo build; the Anytype step docs carry the full instructions, this file carries only position."
user-invocable: true
allowed-tools: Read Write Edit Glob
---

Two modes; pick by intent.

## Resume

Read [STATUS.md](STATUS.md). It is the single source of truth for
cross-thread state: landed decisions (the *why*), uncommitted repo state,
which Anytype step docs are current, and the next numbered steps. Continue
from its **Next** section.

If the working tree disagrees with the file (check `git status`, the step
docs in Anytype), trust the tree — then refresh the file before moving on.

## Compact

Rewrite STATUS.md entirely — never append. Keep it under ~60 lines; details
belong in the Anytype step docs, not here. Sections, in order:

1. **Landed decisions** — one line each, the why not the what
2. **Repo state** — what is uncommitted, what builds/vets clean
3. **Anytype** — which step docs were touched; they hold the instructions
4. **Next** — numbered, smallest next action first

Date-stamp the header so staleness is visible at a glance.
