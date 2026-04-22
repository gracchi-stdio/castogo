# Shoelace Drawer, Menu, Menu Item, Menu Label, Details

## `<sl-drawer>`

Slides in from container edge. Properties: `placement` (start|end|top|bottom), `open`, `contained`, `label`, `no-modal`.
Methods: `show()`, `hide()`.
Events: `sl-show`, `sl-after-show`, `sl-hide`, `sl-after-hide`, `sl-request-close`, `sl-initial-focus`.
Slots: `header-actions`, `footer`, default (body).
Custom property: `--size` (width/height depending on placement).

```html
<sl-drawer label="Menu" placement="start">
  Content here
</sl-drawer>
```

### Contained drawer
Add `contained` attribute + `position: relative` on parent. Contained drawers are not modal (no overlay, no focus trap, no escape dismiss).

### Mobile sidebar pattern
```html
<sl-drawer label="Navigation" placement="start" class="admin-drawer">
  <nav>
    <a href="/admin">Dashboard</a>
    <a href="/admin/episodes">Episodes</a>
  </nav>
</sl-drawer>
<sl-icon-button name="list" label="Menu" onclick="this.previousElementSibling.show()"></sl-icon-button>
```

## `<sl-menu>`, `<sl-menu-item>`, `<sl-menu-label>`

**Important**: Shoelace docs say menus are for *system menus* (dropdowns, context menus). For **navigation**, use `<nav>` and `<a>` elements instead. We use `<nav>` + `<a>` + `<sl-icon>` for sidebar.

Menu items support `prefix`/`suffix` slots for icons/badges:
```html
<sl-menu-item>
  <sl-icon slot="prefix" name="house"></sl-icon>
  Home
</sl-menu-item>
```

`<sl-menu-label>` groups items. `<sl-divider>` separates sections.

## `<sl-details>`

Collapsible sections. Properties: `summary`, `open`, `disabled`, `size`.
Events: `sl-show`, `sl-after-show`, `sl-hide`, `sl-after-hide`.
Slots: `expand-icon`, `collapse-icon`, default (body).
Parts: `summary-icon` (disable animation with `rotate: none`).

Accordion pattern: listen for `sl-show` on container, close all other details.
