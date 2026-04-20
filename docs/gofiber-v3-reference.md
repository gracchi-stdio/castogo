# GoFiber v3 Reference

Compiled reference for the Podlog project. Based on GoFiber v3.1.0.

## 1. Installation & Setup

```bash
go get github.com/gofiber/fiber/v3
```

Requires Go 1.25+.

```go
package main

import (
    "log"
    "github.com/gofiber/fiber/v3"
)

func main() {
    app := fiber.New()

    app.Get("/", func(c fiber.Ctx) error {
        return c.SendString("Hello, World!")
    })

    log.Fatal(app.Listen(":3000"))
}
```

## 2. fiber.Ctx Interface

The context object is an interface for accessing request/response data. Values are reused for performance.

| Method | Purpose |
|--------|---------|
| `Params(key string)` | URL parameters (`:id`) |
| `Query(key string, default string)` | Query parameters |
| `QueryInt(key string, default int)` | Integer query parameters |
| `Get(header string)` | Request header |
| `Set(header string, value string)` | Response header |
| `Cookies(name string, default string)` | Get cookie |
| `Cookie(*fiber.Cookie)` | Set cookie |
| `Bind()` | Body binding methods |
| `UserContext()` | `context.Context` for async |
| `SendString(body string)` | Text response |
| `JSON(body interface{})` | JSON response |
| `Status(code int)` | HTTP status code |
| `SendFile(path string)` | File response |
| `Locals(key string)` | Get request-scoped value |
| `Next()` | Continue to next middleware |
| `Drop()` | Silently close connection |
| `Method()` | HTTP method |
| `Path()` | Request path |
| `IP()` | Client IP |
| `Hostname()` | Request hostname |
| `HasBody()` | Check if request has body |

## 3. Routing

### HTTP Methods

```go
app.Get(path, handler)
app.Post(path, handler)
app.Put(path, handler)
app.Delete(path, handler)
app.Patch(path, handler)
app.All(path, handler)
```

### Route Parameters

```go
app.Get("/users/:id", func(c fiber.Ctx) error {
    id := c.Params("id")
    return c.SendString("User: " + id)
})

app.Get("/users/:id/posts/:postId", func(c fiber.Ctx) error {
    userID := c.Params("id")
    postID := c.Params("postId")
    return c.JSON(fiber.Map{"user": userID, "post": postID})
})
```

### Query Parameters

```go
app.Get("/search", func(c fiber.Ctx) error {
    query := c.Query("q", "")       // with default
    page := c.QueryInt("page", 1)   // as int
    return c.JSON(fiber.Map{"search": query, "page": page})
})
```

## 4. Route Groups

```go
api := app.Group("/api")

v1 := api.Group("/v1")
v1.Get("/users", handler)    // /api/v1/users
v1.Get("/posts", handler)    // /api/v1/posts

v2 := api.Group("/v2")
v2.Get("/users", handler)    // /api/v2/users
```

### Group with Middleware

```go
api := app.Group("/api", func(c fiber.Ctx) error {
    c.Set("X-API-Version", "1.0")
    return c.Next() // MUST call Next()
})

v1 := api.Group("/v1", func(c fiber.Ctx) error {
    c.Set("Version", "v1")
    return c.Next()
})

v1.Get("/list", handler) // /api/v1/list
```

### Multiple Middleware per Group

```go
api.Use(
    func(c fiber.Ctx) error {
        c.Set("X-Custom-Header", "value1")
        return c.Next()
    },
    func(c fiber.Ctx) error {
        c.Set("X-Another-Header", "value2")
        return c.Next()
    },
)
```

## 5. Middleware

### Global

```go
app.Use(func(c fiber.Ctx) error {
    c.Set("X-Custom-Header", "Hello, World")
    return c.Next()
})
```

### Prefix-based

```go
app.Use("/api", func(c fiber.Ctx) error {
    return c.Next()
})

app.Use([]string{"/api", "/admin"}, func(c fiber.Ctx) error {
    return c.Next()
})
```

### Express-style Handlers

```go
app.Use(func(req fiber.Req, res fiber.Res, next func() error) error {
    if req.IP() == "192.168.1.254" {
        return res.SendStatus(fiber.StatusForbidden)
    }
    return next()
})
```

### Execution Order

Middleware runs in registration order. Use `c.Next()` to pass control forward.

```go
app.Get("/",
    func(c fiber.Ctx) error {
        fmt.Println("1st handler!")
        return c.Next()
    },
    func(c fiber.Ctx) error {
        fmt.Println("2nd handler!")
        return c.Next()
    },
    func(c fiber.Ctx) error {
        fmt.Println("3rd handler!")
        return c.SendString("Hello!")
    },
)
```

## 6. Request Handling

### JSON Body Binding (v3 style)

```go
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

app.Post("/users", func(c fiber.Ctx) error {
    user := new(User)
    if err := c.Bind().Body(user); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": err.Error(),
        })
    }
    return c.Status(fiber.StatusCreated).JSON(user)
})
```

### JSON-Only Binding

```go
if err := c.Bind().JSON(&user); err != nil {
    return err
}
```

### Multiple Content-Type Tags

```go
type Person struct {
    Name string `json:"name" xml:"name" form:"name" msgpack:"name"`
    Pass string `json:"pass" xml:"pass" form:"pass" msgpack:"pass"`
}
// Supports JSON, XML, Form-encoded, MessagePack based on Content-Type
```

### Headers and Cookies

```go
// Read
token := c.Get("Authorization")
sessionID := c.Cookies("session_id", "default_value")

// Write
c.Set("X-Custom-Header", "value")
c.Cookie(&fiber.Cookie{
    Name:     "session_id",
    Value:    "abc123",
    Expires:  time.Now().Add(24 * time.Hour),
    HTTPOnly: true,
    Secure:   true,
    SameSite: "Lax",
})
```

### Request Info

```go
ip := c.IP()
method := c.Method()
path := c.Path()
hostname := c.Hostname()
```

## 7. Response Handling

```go
// Text
c.SendString("Hello")

// JSON
c.JSON(fiber.Map{"name": "John", "age": 30})
c.SendJSON(fiber.Map{"name": "John"}) // equivalent alias

// Status code (chainable)
c.Status(fiber.StatusCreated).JSON(user)
c.Status(204).SendString("")

// File
c.SendFile("./files/document.pdf")

// Custom headers
c.Set("Content-Type", "application/json")
c.Set("X-Custom-Header", "Custom Value")
```

## 8. Error Handling

### Return Errors from Handlers

```go
app.Get("/divide/:a/:b", func(c fiber.Ctx) error {
    a, _ := strconv.Atoi(c.Params("a"))
    b, _ := strconv.Atoi(c.Params("b"))

    if b == 0 {
        return fiber.NewError(fiber.StatusBadRequest, "Division by zero")
    }

    return c.JSON(fiber.Map{"result": a / b})
})
```

### Pass Error to Next Handler

```go
app.Get("/", func(c fiber.Ctx) error {
    err := c.SendFile("file-does-not-exist")
    if err != nil {
        return c.Next(err)
    }
    return nil
})
```

### Custom Error Handler

```go
app.SetErrorHandler(func(c fiber.Ctx, err error) error {
    code := fiber.StatusInternalServerError

    var e *fiber.Error
    if errors.As(err, &e) {
        code = e.Code
    }

    return c.Status(code).JSON(fiber.Map{
        "error":  err.Error(),
        "status": code,
    })
})
```

### Error Handler with Content Negotiation

```go
app.SetErrorHandler(func(c fiber.Ctx, err error) error {
    code := fiber.StatusInternalServerError
    var e *fiber.Error
    if errors.As(err, &e) {
        code = e.Code
    }

    if c.Accepts("json") == "json" || c.Get("Content-Type") == "application/json" {
        return c.Status(code).JSON(fiber.Map{"error": err.Error(), "status": code})
    }

    return c.Status(code).SendString(err.Error())
})
```

### Built-in Error Sentinels

```go
fiber.ErrBadRequest
fiber.ErrNotFound
fiber.ErrMethodNotAllowed
fiber.ErrInternalServerError
fiber.ErrUnauthorized
fiber.ErrForbidden
fiber.ErrConflict

// Custom fiber error
err := fiber.NewError(fiber.StatusBadRequest, "Custom message")
```

## 9. Async Operations

### UserContext

```go
app.Get("/async", func(c fiber.Ctx) error {
    ctx := c.UserContext() // context.Context
    go performAsyncTask(ctx)
    return c.SendString("Task started")
}
```

### Timeout Pattern

```go
app.Get("/timeout-task", func(c fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
    defer cancel()

    result := make(chan string)

    go func() {
        time.Sleep(1 * time.Second)
        result <- "done"
    }()

    select {
    case res := <-result:
        return c.JSON(fiber.Map{"status": res})
    case <-ctx.Done():
        return fiber.NewError(fiber.StatusRequestTimeout, "Operation timeout")
    }
})
```

## 10. Complete CRUD Example

```go
package main

import (
    "errors"
    "log"
    "strconv"

    "github.com/gofiber/fiber/v3"
)

type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

var users = []User{
    {ID: 1, Name: "Alice", Email: "alice@example.com"},
    {ID: 2, Name: "Bob", Email: "bob@example.com"},
}

func main() {
    app := fiber.New()

    app.Use(func(c fiber.Ctx) error {
        log.Printf("%s %s", c.Method(), c.Path())
        return c.Next()
    })

    api := app.Group("/api")
    v1 := api.Group("/v1", func(c fiber.Ctx) error {
        c.Set("X-API-Version", "1.0")
        return c.Next()
    })

    // GET all
    v1.Get("/users", func(c fiber.Ctx) error {
        return c.JSON(users)
    })

    // GET by ID
    v1.Get("/users/:id", func(c fiber.Ctx) error {
        id, _ := strconv.Atoi(c.Params("id"))
        for _, user := range users {
            if user.ID == id {
                return c.JSON(user)
            }
        }
        return fiber.NewError(fiber.StatusNotFound, "User not found")
    })

    // POST create
    v1.Post("/users", func(c fiber.Ctx) error {
        user := new(User)
        if err := c.Bind().Body(user); err != nil {
            return fiber.NewError(fiber.StatusBadRequest, err.Error())
        }
        users = append(users, *user)
        return c.Status(fiber.StatusCreated).JSON(user)
    })

    // Custom error handler
    app.SetErrorHandler(func(c fiber.Ctx, err error) error {
        code := fiber.StatusInternalServerError
        message := err.Error()

        var e *fiber.Error
        if errors.As(err, &e) {
            code = e.Code
            message = e.Message
        }

        return c.Status(code).JSON(fiber.Map{"error": message})
    })

    log.Fatal(app.Listen(":3000"))
}
```

## 11. HTTP Status Codes

```go
// 2xx Success
fiber.StatusOK                   // 200
fiber.StatusCreated              // 201
fiber.StatusAccepted             // 202
fiber.StatusNoContent            // 204

// 3xx Redirection
fiber.StatusMovedPermanently     // 301
fiber.StatusFound                // 302
fiber.StatusNotModified          // 304

// 4xx Client Error
fiber.StatusBadRequest           // 400
fiber.StatusUnauthorized         // 401
fiber.StatusForbidden            // 403
fiber.StatusNotFound             // 404
fiber.StatusMethodNotAllowed     // 405
fiber.StatusConflict             // 409
fiber.StatusGone                 // 410

// 5xx Server Error
fiber.StatusInternalServerError  // 500
fiber.StatusNotImplemented       // 501
fiber.StatusBadGateway           // 502
fiber.StatusServiceUnavailable   // 503
```

## 12. v2 to v3 Migration

| v2 | v3 |
|----|-----|
| `*fiber.Ctx` | `fiber.Ctx` (interface, value receiver) |
| `c.BodyParser(&x)` | `c.Bind().Body(&x)` |
| `c.BodyParser(&x)` (JSON only) | `c.Bind().JSON(&x)` |
| `app.Settings.ErrorHandler` | `app.SetErrorHandler()` |
| Various context methods | Improved Ctx interface |

### Common Mistakes

| Wrong | Correct |
|-------|---------|
| `*fiber.Ctx` | `fiber.Ctx` |
| `c.BodyParser()` | `c.Bind().Body()` |
| `app.Settings.ErrorHandler` | `app.SetErrorHandler()` |
| Middleware without `c.Next()` | Always call `c.Next()` |
| `import "fiber"` | `import "github.com/gofiber/fiber/v3"` |

## 13. Testing with cURL

```bash
# GET
curl http://localhost:3000/api/v1/users

# GET with query params
curl "http://localhost:3000/api/v1/users?page=1&limit=10"

# POST with JSON
curl -X POST http://localhost:3000/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"id":3,"name":"Charlie","email":"charlie@example.com"}'

# With auth header
curl -H "Authorization: Bearer token123" \
  http://localhost:3000/api/v1/users

# With cookie
curl -H "Cookie: session_id=abc123" \
  http://localhost:3000/api/v1/users
```
