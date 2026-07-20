---
title: Pages & Blocks
description: Build site pages from composable content blocks.
order: 2
---

# Pages & Blocks

Your public site is made of **pages**, and each page is assembled from ordered
**blocks** (a hero, a feature grid, an episode list, etc.). This guide walks
through creating a page and filling it with blocks.

## Create a page

1. In the sidebar, click **Pages**. You'll see the pages table
   (Title · Path · Status · Actions).
2. Click **New Page** (top-right).
3. Fill in the form:

   | Field | What it does |
   |-------|--------------|
   | **Title** | The page name (required). |
   | **Slug** | The URL segment — the page lives at `/{slug}`. **Leave blank for your homepage** (see below). |
   | **Parent Page** | Nest this under another page to get a URL like `/about/team`. Defaults to *None (top-level)*. |
   | **Published** | Off by default — the page starts as a draft and is **not** visible publicly until you check this. |
   | **Show in main navigation** | Show a link to this page in the site navbar. |

4. Click **Save Changes**. You're taken to the new page's **Settings** tab.

### Reserved slugs

You can't use these as a slug (they're reserved for the app): `admin`, `api`,
`login`, `signup`, `register`, `healthcheck`, `logout`, `feed`, `assets`.

### Your homepage

There is no separate homepage picker. The homepage is **the one top-level page
with an empty slug** — its public URL is `/`. Only one such page can exist; if
you already have one, the Slug field must be filled in.

### Nesting

Pages can be nested **one level deep** (parent → child), giving URLs like
`/about/team`. Deeper nesting isn't supported.

## Add blocks

1. On the page, click the **Blocks** tab in the subnav.
2. The editor has two panes: the **block list** (left) and the **block editor**
   (right).
3. To add a block, choose a type from the dropdown and click **Add Block**. The
   new block appears in the list **and** opens immediately in the right pane for
   editing.

### Arrange blocks

- **Reorder**: drag a block by its grip handle (left side of the card).
- **Edit**: click a block in the list to load it in the right pane, change its
  fields, then click **Save Changes**.
- **Delete**: click the trash icon on a block's card.

## Block types

| Type | Use it for |
|------|------------|
| **Hero** | Top banner: headline, subheadline, call-to-action, background image. |
| **Features** | A grid of items, each with an icon, title, and description. |
| **Episodes** | Automatically lists your latest published episodes. |
| **Testimonials** | Quotes from listeners, each with author/role/avatar. |
| **Call to Action** | A focused headline + description + button. |
| **Footer** | Copyright, navigation links, and social links. |
| **Rich Text** | Free-form markdown content. |

### Repeatable items

**Features**, **Testimonials**, and **Footer** contain lists of items. Inside
those block editors, use the **Add Feature** / **Add Testimonial** / **Add
Link** / **Add Social** buttons to append items, and each item's trash icon to
remove it.

### Rich Text blocks

The **Rich Text** block accepts markdown — headings, lists, tables, code blocks,
links. It renders with the site's typography styling on the public page.

## Publish the page

A page is only visible on the public site when it's published.

1. Open the **Settings** tab.
2. Check **Published**.
3. Click **Save Changes**.

The page is now live at `/{slug}` (or `/` for the homepage). Unpublished pages
return a 404 to public visitors.
