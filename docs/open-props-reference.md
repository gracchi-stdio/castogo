# Open Props Reference for Podlog

## Setup (CDN)

```html
<!-- Base tokens (required) -->
<link rel="stylesheet" href="https://unpkg.com/open-props" />

<!-- Optional extras -->
<link rel="stylesheet" href="https://unpkg.com/open-props/normalize.min.css" />
<link rel="stylesheet" href="https://unpkg.com/open-props/buttons.min.css" />
```

Only the base import is needed for design tokens. Don't import normalize/buttons until ready — they override existing styles.

## Color System

### Static Colors (oklch)

19 color scales, each 0-12 (0=lightest, 12=darkest):

```
--gray-{0-12}    --red-{0-12}     --pink-{0-12}
--purple-{0-12}  --violet-{0-12}  --indigo-{0-12}
--blue-{0-12}    --cyan-{0-12}    --teal-{0-12}
--green-{0-12}   --lime-{0-12}    --yellow-{0-12}
--orange-{0-12}  --choco-{0-12}   --brown-{0-12}
--sand-{0-12}    --camo-{0-12}    --jungle-{0-12}
--stone-{0-12}
```

### Dynamic Palette (oklch, any hue)

```css
@import "open-props/palette";

:root {
  --palette-hue: 270;         /* 0-360 */
  --palette-chroma: 0.9;      /* saturation */
  --palette-hue-rotate-by: 1; /* step between shades */
}
/* Then use --color-{1-16} */
```

### Adaptive Theme Colors

```css
/* Light (default) */
:root {
  --text-1: var(--gray-8);        /* primary text */
  --text-2: var(--gray-7);        /* secondary text */
  --surface-1: var(--gray-0);     /* page background */
  --surface-2: var(--gray-1);     /* card background */
  --surface-3: var(--gray-2);     /* elevated surface */
  --surface-4: var(--gray-3);     /* hover/active */
  --link: var(--indigo-7);
  --brand: var(--orange-6);
}

/* Dark */
@media (prefers-color-scheme: dark) {
  :root {
    --text-1: var(--gray-3);
    --text-2: var(--gray-5);
    --surface-1: var(--gray-12);
    --surface-2: var(--gray-11);
    --surface-3: var(--gray-10);
    --surface-4: var(--gray-9);
    --brand: var(--orange-3);
  }
}
```

### Theme Switching (user toggle)

```html
<link rel="stylesheet" href="https://unpkg.com/open-props/theme.light.switch.min.css" />
<link rel="stylesheet" href="https://unpkg.com/open-props/theme.dark.switch.min.css" />
```

Enables selectors: `.dark`, `.light`, `[data-theme="dark"]`, `[data-theme="light"]`

## Typography

### Font Sizes

```
--font-size-00: .5rem    →  --font-size-8: 3.5rem
--font-size-fluid-0: clamp(.75rem, 2vw, 1rem)
--font-size-fluid-3: clamp(2rem, 9vw, 3.5rem)
```

### Font Families (system fonts, no external loading)

```
--font-sans               → system-ui
--font-monospace-code     → SF Mono, Cascadia Code, Consolas...
--font-transitional       → Charter, Cambria...
--font-geometric-humanist → Avenir, Montserrat...
```

### Font Weights, Line Heights, Letter Spacing

```
--font-weight-{1-9}           → 100 to 900
--font-lineheight-{00-5}      → .95 to 2
--font-letterspacing-{0-7}    → -.05em to 1em
```

## Spacing

```
--size-{000-15}           → -.5rem to 30rem (rem-based)
--size-px-{000-15}        → -8px to 480px (px-based)
--size-fluid-{1-10}       → clamp-based responsive
--size-content-{1-3}      → 20ch, 45ch, 60ch (reading widths)
--size-header-{1-3}       → 20ch, 25ch, 35ch (headline widths)
```

Common: `--size-1` (.25rem), `--size-2` (.5rem), `--size-3` (1rem), `--size-4` (1.25rem), `--size-5` (1.5rem)

## Borders & Radius

```
--border-size-{1-5}             → 1px to 5px
--radius-{1-6}                  → subtle to very round
--radius-round                  → 9999px (pill/circle)
--radius-conditional-{1-6}      → no radius when fullscreen
--radius-blob-{1-5}             → organic blob shapes
--radius-drawn-{1-6}            → hand-drawn look
```

## Shadows

```
--shadow-{1-6}         → light to heavy outer shadows
--inner-shadow-{0-4}   → inner shadows
```

## Easing

```
--ease-{1-5}                  → standard ease strengths
--ease-in-{1-5}               → ease-in strengths
--ease-out-{1-5}              → ease-out strengths
--ease-elastic-out-{1-5}      → elastic/bouncy
--ease-spring-{1-5}           → spring physics
--ease-bounce-{1-5}           → bounce effect
```

## Animations

Premade keyframe effects:

```
--animation-fade-{in,out}
--animation-slide-in-{up,down,left,right}
--animation-slide-out-{up,down,left,right}
--animation-spin, --animation-pulse, --animation-bounce
--animation-shake-{x,y}
--animation-float, --animation-blink, --animation-ping
```

Can be combined:
```css
.element {
  animation:
    var(--animation-fade-in),
    var(--animation-slide-in-up);
  animation-duration: var(--duration-moderate-2);
  animation-timing-function: var(--ease-3);
}
```

## Media Queries (PostCSS or custom-media)

```css
@media (--OSdark) { }      → prefers-color-scheme: dark
@media (--OSlight) { }     → prefers-color-scheme: light
@media (--motionOK) { }    → prefers-reduced-motion: no-preference
@media (--md-n-above) { }  → width >= 768px
@media (--touch) { }       → touch device
@media (--mouse) { }       → mouse device
```

## Gradients

```
--gradient-{1-30}    → 30 handcrafted gradients
--noise-{1-5}        → grainy noise textures
```

## Z-Index

```
--layer-{1-5}            → 1 to 5
--layer-important        → 2147483647 (max)
```

## Shoelace Integration

Override Shoelace tokens with Open Props values in `app.css`:

```css
:root {
  --sl-color-primary-700: var(--indigo-7);
  --sl-color-primary-600: var(--indigo-6);
  --sl-color-neutral-700: var(--gray-7);
  --sl-input-border-color: var(--gray-6);
  --sl-border-radius-medium: var(--radius-2);
  --sl-font-size-medium: var(--font-size-2);
}
```

Shoelace has its own design tokens (`--sl-*`). Open Props provides richer tokens. Use Open Props as the source and map to Shoelace where needed.
