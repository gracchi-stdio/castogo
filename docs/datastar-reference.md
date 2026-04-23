# Datastar Reference for Podlog

## SDK

- Go SDK: `github.com/starfederation/datastar-go/datastar`
- Client runtime: loaded via CDN in `internal/view/base_layout.templ`

## Philosophy

> Why should only `<a>` & `<form>` be able to make HTTP requests?
> Why should only `click` & `submit` events trigger them?
> Why should only `GET` & `POST` methods be available?
> Why should you only be able to replace the **entire** screen?

Datastar removes these constraints. The backend drives the frontend by sending HTML patches. Signals handle user interactions. No SPA framework needed.

**Two capabilities:**
1. **Backend reactivity** — server patches DOM via SSE
2. **Frontend reactivity** — `data-*` attributes for client-side state (like Alpine.js)

**Backend is source of truth.** Signals are for: capturing user input, temporary UI state, sending state to backend.

## Core Flow

1. Client declares reactive state: `data-signals:delay="'400'"`
2. Input binds to state: `data-bind:delay`
3. Action triggers backend: `data-on:click="@get('/hello-world')"`
4. Server reads signals: `datastar.ReadSignals(r, &store)`
5. Server opens SSE: `datastar.NewSSE(w, r)`
6. Server patches DOM: `sse.PatchElements(html)`

Datastar morphs new HTML into the existing DOM — preserves scroll position, focus, etc. Only updates what changed. Elements need stable `id` attributes for morphing to work.

## Signal Transmission

Signals are sent to the server automatically with every backend request:
- **GET/DELETE**: signals sent as `datastar` query parameter
- **Other methods**: signals sent as JSON in request body

Signals starting with `_` are excluded from transmission (use for private client state).

## Client-Side Attributes

### State

| Attribute | Purpose | Example |
|---|---|---|
| `data-signals:*` | Declare reactive state | `data-signals:count="0"` |
| `data-bind:*` | Two-way bind input to signal | `<input data-bind:delay>` |
| `data-computed:*` | Derived/read-only signal | `data-computed:doubled="$value * 2"` |

### Display

| Attribute | Purpose | Example |
|---|---|---|
| `data-text` | Set element text from expression | `<div data-text="$count">` |
| `data-show` | Show/hide based on condition | `data-show="$count > 0"` |
| `data-class:*` | Toggle CSS class | `data-class:active="$isActive"` |
| `data-attr:*` | Bind any HTML attribute | `data-attr:disabled="$isDisabled"` |

### Actions

| Attribute | Purpose | Example |
|---|---|---|
| `data-on:click` | Action on click | `data-on:click="@get('/path')"` |
| `data-on:submit` | Action on form submit | `data-on:submit="@post('/path')"` |
| `data-on:sl-change` | Shoelace select change | `data-on:sl-change="$val = el.value"` |
| `data-init` | Auto-trigger on page load | `data-init="@get('/stream')"` |

Built-in actions: `@get`, `@post`, `@put`, `@patch`, `@delete`.

### SSE Action Options

```html
<div data-init="@get('/stream', {retryInterval: 100, retry: 'always'})">
```

- `retryInterval` — ms between reconnection attempts
- `retry: 'always'` — keep retrying after successful connection

Custom headers in actions:
```html
<button data-on:click="@post('/api', {
    headers: { 'Authorization': 'Bearer ' + $token }
})">
```

## Expressions

Datastar expressions are JavaScript evaluated in `data-*` attributes. Signals accessed via `$` prefix.

```html
<!-- Arithmetic -->
<div data-text="$count + 1"></div>

<!-- Conditional -->
<div data-show="$age >= 18">Adult</div>

<!-- Multiple statements (semicolons) -->
<button data-on:click="$count++; $total = $count * $price">Add</button>

<!-- Function calls -->
<div data-text="$name.toUpperCase()"></div>

<!-- Set signal value -->
<button data-on:click="$hal = 'I read you, Dave.'">Ping</button>
```

## Patching Modes

Control how Datastar merges HTML via response headers:

| Header | Values | Default |
|---|---|---|
| `datastar-selector` | CSS selector for target elements | — |
| `datastar-mode` | `outer`, `inner`, `remove`, `replace`, `prepend`, `append`, `before`, `after` | `outer` |
| `datastar-use-view-transition` | `true` to use View Transition API | — |

Example (non-Go server):
```
response.headers.set('datastar-mode', 'inner')
response.headers.set('datastar-use-view-transition', 'true')
```

## Server-Side API

### Reading Signals

```go
store := &MySignals{}
if err := datastar.ReadSignals(r, store); err != nil {
    // handle error
}
```

- GET/DELETE: signals from URL query param `datastar`
- Other methods: signals from request body
- Target struct uses `json` tags

### SSE + DOM Patching

```go
sse := datastar.NewSSE(w, r)
sse.PatchElements(`<div id="msg">Hello</div>`)
```

### Patching Signals from Backend

```go
sse.PatchSignals(`{"hal": "Affirmative, Dave."}`)
sse.MarshalAndPatchSignals(myStruct)              // convenience helper
sse.MarshalAndPatchSignalsIfMissing(myStruct)      // only sets missing keys
```

### Executing Client-Side JavaScript

```go
sse.ExecuteScript("window.location.reload()")
```

### Long-Lived SSE Connections

```go
sse := datastar.NewSSE(w, r)
// ... send initial patches ...
<-r.Context().Done() // hold until client disconnects
```

### Templ Component Streaming

```go
sse.PatchElementTempl(ctx, myTemplComponent())
```

Adopt after literal HTML strings work — debug one thing at a time.

## Echo + Datastar Pattern

Datastar's SDK calls `http.ResponseController.Flush()` internally, which conflicts with Echo's `Response.Flush()` (calls `WriteHeader` if not committed). Fix: pass the **raw writer** `c.Response().Writer`.

Shared helpers (in `internal/handler/helpers.go`):

```go
func sse(c echo.Context) *datastar.ServerSentEventGenerator {
    return datastar.NewSSE(c.Response().Writer, c.Request())
}

func readSignals(c echo.Context, target any) error {
    return datastar.ReadSignals(c.Request(), target)
}
```

Usage:

```go
// One-shot SSE (login, form submit)
sse(c).MarshalAndPatchSignals(map[string]string{"error": "Invalid"})
sse(c).Redirect("/admin")

// Long-lived SSE (live updates)
out := sse(c)
for {
    select {
    case <-ticker.C:
        out.PatchElements(`<div id="clock">` + now + `</div>`)
    case <-c.Request().Context().Done():
        return nil
    }
}

// Reading signals from frontend
input := &MySignals{}
if err := readSignals(c, input); err != nil {
    sse(c).MarshalAndPatchSignals(map[string]string{"error": "Invalid request"})
    return nil
}
```

Session before SSE — `sess.Save()` adds Set-Cookie header, then `sse(c)` commits everything:

```go
sess.Save(c.Request(), c.Response().Writer)
sse(c).Redirect("/admin")
```

## Animations & View Transitions

### Technique 1: Stable ID + CSS Transitions

Keep element IDs stable across SSE swaps. CSS transitions apply automatically:

```css
#color-throb { transition: all 0.5s ease; }
```

### Technique 2: View Transitions API

Works automatically with Datastar's SSE swapping. Not yet in Firefox.

```css
::view-transition-old(#list) { animation: fade-out 0.2s ease; }
::view-transition-new(#list) { animation: fade-in 0.2s ease; }
```

### Technique 3: Fade Out on Removal

Send opacity 0 before removing via SSE:
```html
<div id="item" style="opacity: 0; transition: opacity 0.3s;">
```

### Technique 4: Fade In on Addition

```css
.fade-in { opacity: 0; transition: opacity 0.3s; }
```

## Development: Hot Reload Pattern

```html
<body data-init="@get('/hotreload', {retryInterval: 100, retry: 'always'})">
```

```go
var once sync.Once

func HotReloadHandler(w http.ResponseWriter, r *http.Request) {
    sse := datastar.NewSSE(w, r)
    once.Do(func() {
        sse.ExecuteScript("window.location.reload()")
    })
    <-r.Context().Done()
}
```

Pair with [Reflex](https://github.com/cespare/reflex) for file-watcher restarts.

## Gotchas

- **Signal values are JS expressions**: `data-signals:status="all"` evaluates `all` as variable — use `data-signals:status="'all'"` (quoted string).
- **Shoelace web components**: `data-bind` does NOT work (they fire `sl-change` not `input`). Use `data-on:sl-change="$signal = el.value"`.
- **Stable IDs matter**: morphing needs stable element IDs to diff correctly.
- **Backend drives UI**: if you're building complex frontend signal state, move logic to the backend.

## Adoption Order

1. Literal HTML strings with `PatchElements` (hello-world parity)
2. `PatchElementTempl` for templ component streaming
3. `PatchSignals` / `MarshalAndPatchSignals` for pushing state updates
4. View transitions and animations

## Verification Checklist

1. Page source includes Datastar script tag
2. `make generate` was run and generated files match source templates
3. SSE response stream includes Datastar patch events (check browser Network tab)
4. Target element updates without full page reload
5. Bound input values are sent to backend in signal payload
