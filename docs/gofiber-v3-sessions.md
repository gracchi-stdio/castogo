# GoFiber v3 Session Middleware Reference

**Podlog note:** This project uses DB-backed sessions stored in PostgreSQL (sessions table with token + expires_at). The session queries are in `sql/queries/` and generated code in `internal/db/`.

## Quick Start

```go
import (
    "github.com/gofiber/fiber/v3"
    "github.com/gofiber/fiber/v3/middleware/session"
)

app.Use(session.New())

app.Get("/", func(c fiber.Ctx) error {
    sess := session.FromContext(c)

    visits := 1
    if v := sess.Get("visits"); v != nil {
        if vInt, ok := v.(int); ok {
            visits = vInt + 1
        }
    }
    sess.Set("visits", visits)

    return c.SendString(fmt.Sprintf("Visits: %d", visits))
})
```

## Session Methods

### Getting Data

```go
sess := session.FromContext(c)

// Get with type assertion
if value := sess.Get("key"); value != nil {
    if strVal, ok := value.(string); ok {
        // Use strVal
    }
}

keys := sess.Keys()      // All keys
id := sess.ID()          // Session ID
```

### Setting Data

```go
sess := session.FromContext(c)

sess.Set("user_id", 123)
sess.Set("username", "john")
sess.Set("roles", []string{"admin", "user"})
sess.SetExpiry(2 * time.Hour)
```

### Deleting / Resetting

```go
sess.Delete("user_id")              // Delete single key

if err := sess.Reset(); err != nil { // Clear all, new ID
    return c.Status(500).SendString(err.Error())
}

if err := sess.Destroy(); err != nil { // Destroy completely
    return c.Status(500).SendString(err.Error())
}

if err := sess.Regenerate(); err != nil { // Keep data, new ID
    return c.Status(500).SendString(err.Error())
}
```

### Manual Save

```go
sess.Set("data", "value")
if err := sess.Save(); err != nil {
    return c.Status(500).SendString(err.Error())
}
```

## Configuration

### Full Config Reference

| Option | Type | Description | Default |
|--------|------|-------------|---------|
| `Storage` | `fiber.Storage` | Storage backend (memory, Redis, PostgreSQL, etc.) | `memory.New()` |
| `IdleTimeout` | `time.Duration` | Inactivity timeout | `24 * time.Hour` |
| `AbsoluteTimeout` | `time.Duration` | Maximum session lifetime | infinite |
| `CookieName` | `string` | Session cookie name | `session_id` |
| `CookieDomain` | `string` | Cookie domain | `` |
| `CookiePath` | `string` | Cookie path | `` |
| `CookieSecure` | `bool` | HTTPS only | `false` |
| `CookieHTTPOnly` | `bool` | Prevent JS access | `false` |
| `CookieSameSite` | `string` | CSRF protection (Lax, Strict, None) | `Lax` |
| `CookieSessionOnly` | `bool` | Browser session only | `false` |
| `KeyLookup` | `string` | Where to get session ID | `cookie:session_id` |
| `KeyGenerator` | `func() string` | Generate session ID | `utils.UUIDv4` |
| `Extractor` | `extractors.Extractor` | Custom extraction | Cookie-based |

### Production Setup with Redis

```go
import (
    "time"
    "github.com/gofiber/fiber/v3/middleware/session"
    "github.com/gofiber/storage/redis"
    "github.com/gofiber/fiber/v3/extractors"
)

redisStorage := redis.New(redis.Config{
    Host:     "localhost",
    Port:     6379,
    Password: "",
    Database: 0,
})

app.Use(session.New(session.Config{
    Storage:           redisStorage,
    CookieSecure:      true,
    CookieHTTPOnly:    true,
    CookieSameSite:    "Lax",
    IdleTimeout:       30 * time.Minute,
    AbsoluteTimeout:   24 * time.Hour,
    Extractor:         extractors.FromCookie("__Host-session_id"),
}))
```

## Session Extractors

```go
import "github.com/gofiber/fiber/v3/extractors"

extractors.FromCookie("session_id")     // From cookie (default)
extractors.FromHeader("X-Session-Token") // From header
extractors.FromQuery("session_id")       // From query parameter
extractors.FromForm("session_id")        // From form
```

## Storage Backends

### In-Memory (Development Only)

```go
app.Use(session.New())
// Warning: Sessions lost on restart, single-server only
```

### Redis

```bash
go get github.com/gofiber/storage/redis/v3
```

```go
import "github.com/gofiber/storage/redis"

storage := redis.New(redis.Config{
    Host:     "localhost",
    Port:     6379,
    Password: "your_password",
    Database: 0,
})

app.Use(session.New(session.Config{Storage: storage}))
```

### PostgreSQL

```bash
go get github.com/gofiber/storage/postgres
```

```go
import "github.com/gofiber/storage/postgres"

storage := postgres.New(postgres.Config{
    Host:     "localhost",
    Port:     5432,
    Database: "sessions",
    Username: "user",
    Password: "pass",
})

app.Use(session.New(session.Config{Storage: storage}))
```

### SQLite3

```bash
go get github.com/gofiber/storage/sqlite3
```

```go
import "github.com/gofiber/storage/sqlite3"

storage := sqlite3.New(sqlite3.Config{Database: "./sessions.db"})

app.Use(session.New(session.Config{Storage: storage}))
```

Other storage drivers available at `github.com/gofiber/storage/`

## Login/Logout Pattern

```go
// Login - create session
app.Post("/login", func(c fiber.Ctx) error {
    email := c.FormValue("email")
    password := c.FormValue("password")

    if !isValidUser(email, password) {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "Invalid credentials",
        })
    }

    sess := session.FromContext(c)

    // Regenerate ID (prevent fixation attacks)
    if err := sess.Regenerate(); err != nil {
        return c.Status(500).SendString(err.Error())
    }

    sess.Set("user_id", 123)
    sess.Set("email", email)
    sess.Set("authenticated", true)

    return c.JSON(fiber.Map{"status": "logged in"})
})

// Protected route - check session
app.Get("/dashboard", func(c fiber.Ctx) error {
    sess := session.FromContext(c)

    auth := sess.Get("authenticated")
    if auth == nil || auth != true {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "Not authenticated",
        })
    }

    email := sess.Get("email")
    return c.JSON(fiber.Map{"message": "Welcome " + email.(string)})
})

// Logout - destroy session
app.Post("/logout", func(c fiber.Ctx) error {
    sess := session.FromContext(c)
    if err := sess.Reset(); err != nil {
        return c.Status(500).SendString(err.Error())
    }
    return c.JSON(fiber.Map{"status": "logged out"})
})
```

## API Rate Limiting with Sessions

```go
app.Use(session.New(session.Config{
    Extractor: extractors.FromHeader("X-API-Token"),
    Storage:   redis.New(),
}))

app.Post("/api/data", func(c fiber.Ctx) error {
    sess := session.FromContext(c)

    var callCount int
    if cc := sess.Get("api_calls"); cc != nil {
        if count, ok := cc.(int); ok {
            callCount = count
        }
    }

    if callCount > 100 {
        return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
            "error": "Rate limit exceeded",
        })
    }

    callCount++
    sess.Set("api_calls", callCount)
    sess.Set("last_call", time.Now())

    return c.JSON(fiber.Map{"data": "some data", "calls": callCount})
})
```

## Security Best Practices

1. **HTTPS only in production** — set `CookieSecure: true`
2. **HTTPOnly flag** — set `CookieHTTPOnly: true` to prevent XSS JS access
3. **SameSite attribute** — set `CookieSameSite: "Lax"` or `"Strict"`
4. **Regenerate on login** — call `sess.Regenerate()` to prevent fixation attacks
5. **Idle timeout** — set `IdleTimeout: 30 * time.Minute`
6. **Absolute timeout** — set `AbsoluteTimeout: 24 * time.Hour`
7. **Persistent storage** — never use in-memory sessions in production
8. **Validate session data** — always type-assert with `ok` check
9. **Reset on logout** — call `sess.Reset()` to clear all data and generate new ID

```go
// Safe type assertion
userID := sess.Get("user_id")
if userID == nil {
    return c.Status(fiber.StatusUnauthorized).SendString("Not authenticated")
}
uid, ok := userID.(int)
if !ok {
    return c.Status(fiber.StatusInternalServerError).SendString("Invalid session")
}
```

## Complete Production Example

```go
package main

import (
    "fmt"
    "time"

    "github.com/gofiber/fiber/v3"
    "github.com/gofiber/fiber/v3/middleware/session"
    "github.com/gofiber/fiber/v3/extractors"
    "github.com/gofiber/storage/redis"
)

func main() {
    app := fiber.New()

    storage := redis.New(redis.Config{
        Host: "localhost",
        Port: 6379,
    })

    app.Use(session.New(session.Config{
        Storage:         storage,
        CookieSecure:    true,
        CookieHTTPOnly:  true,
        CookieSameSite:  "Lax",
        IdleTimeout:     30 * time.Minute,
        AbsoluteTimeout: 24 * time.Hour,
        Extractor:       extractors.FromCookie("__Host-session_id"),
    }))

    app.Post("/login", login)
    app.Get("/dashboard", authenticate, dashboard)
    app.Get("/profile", authenticate, profile)
    app.Post("/logout", logout)

    app.Listen(":3000")
}

func authenticate(c fiber.Ctx) error {
    sess := session.FromContext(c)
    if sess.Get("authenticated") != true {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "Not authenticated",
        })
    }
    return c.Next()
}

func login(c fiber.Ctx) error {
    email := c.FormValue("email")
    password := c.FormValue("password")

    if email != "user@example.com" || password != "password" {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "error": "Invalid credentials",
        })
    }

    sess := session.FromContext(c)
    sess.Regenerate()
    sess.Set("user_id", 123)
    sess.Set("email", email)
    sess.Set("authenticated", true)

    return c.JSON(fiber.Map{"status": "logged in"})
}

func dashboard(c fiber.Ctx) error {
    sess := session.FromContext(c)
    email := sess.Get("email")
    return c.JSON(fiber.Map{"message": fmt.Sprintf("Welcome %v", email)})
}

func profile(c fiber.Ctx) error {
    sess := session.FromContext(c)
    userID := sess.Get("user_id")
    return c.JSON(fiber.Map{"user_id": userID})
}

func logout(c fiber.Ctx) error {
    sess := session.FromContext(c)
    sess.Reset()
    return c.JSON(fiber.Map{"status": "logged out"})
}
```

## Performance Tips

1. Use Redis for production — in-memory storage doesn't scale
2. Set appropriate timeouts — balance security and UX
3. Type assert safely — always use `ok` check
4. Minimize session reads/writes per request
5. Monitor storage — watch Redis/DB for session growth
