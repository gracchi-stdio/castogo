---
name: Admin UI consistency
description: Use shared components for all admin UI, maintain consistency across admin pages
type: feedback
originSessionId: b903453f-2825-4e5a-a9eb-8fe6360c8e0b
---
Always use the shared component library (`internal/view/components/`) for all admin page development — never write raw HTML form elements. Keep admin pages visually consistent with each other.

**Why:** The project has a shared component system (button, input, select, dialog, etc.) to ensure theme scope changes propagate and visual consistency is maintained. Bypassing it breaks that consistency.

**How to apply:** When building admin pages for Pages CMS (or any feature), import and use `button.Button(...)`, `input.Input(...)`, `select.Select(...)`, etc. from `internal/view/components/`. Check existing admin views for patterns to follow.
