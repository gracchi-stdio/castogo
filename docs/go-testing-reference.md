# Go Testing Reference

## Running Tests

```bash
# All tests in one package
go test ./internal/service/ -v

# Specific test by name (-run accepts a regex)
go test ./internal/service/ -run TestParseLogEntry -v

# All tests in entire project
go test ./... -v

# The -v flag means "verbose" — shows each test name and PASS/FAIL
```

## Test File Conventions

- File must be named `*_test.go` — Go's test runner only finds this pattern
- Same package as the code being tested (e.g., `package service`) — can access unexported functions/types
- Test function must start with `Test` (uppercase) and receive `*testing.T`

```go
package service

import "testing"

func TestSomething(t *testing.T) {
    // test code here
}
```

## Assertions

Go has no built-in assert library. Use `*testing.T` methods directly:

| Method | Behavior | Use when |
|---|---|---|
| `t.Fatalf(msg, args...)` | Logs message, **stops test immediately** | Further assertions would crash or are meaningless (e.g., nil pointer) |
| `t.Errorf(msg, args...)` | Logs message, **test continues** | Checking multiple independent fields — see all failures at once |
| `t.Log(msg, args...)` | Logs message (only shown with `-v`) | Debug output, not a failure |

Common format verbs: `%v` (any value), `%d` (integer), `%s` (string), `%T` (type name).

## Test Patterns

### Happy Path — valid input, expect correct output

```go
func TestParseLogEntry(t *testing.T) {
    entry, err := ParseLogEntry("HIT|200|1507167062421|412|...")

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if entry.StatusCode != 200 {
        t.Errorf("StatusCode: want 200, got %d", entry.StatusCode)
    }
}
```

### Error Path — invalid input, expect an error

```go
func TestParseLogEntry_InvalidLine(t *testing.T) {
    entry, err := ParseLogEntry("too|few|fields")

    if err == nil {
        t.Fatal("expected error for malformed line, got nil")
    }
    if entry != nil {
        t.Error("expected nil entry on error")
    }
}
```

### HTTP Handler — fake server with `httptest.NewServer`

For code that makes HTTP requests, use `net/http/httptest`:

```go
func TestFetcher(t *testing.T) {
    // 1. Create fake server — controls what the "API" responds with
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify the request (headers, path, method)
        if r.Header.Get("AccessKey") != "test-key" {
            t.Errorf("wrong AccessKey: %s", r.Header.Get("AccessKey"))
        }
        // Send response (plain text, gzip, JSON, etc.)
        w.Write([]byte("response data"))
    }))
    defer server.Close()

    // 2. Point your code at the fake server
    fetcher := NewFetcher(server.URL, "test-key", server.Client())

    // 3. Call and assert
    result, err := fetcher.Fetch(context.Background())
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    // check result fields...
}
```

### Gzip Response — simulating compressed API responses

When the real API sends gzip (like Bunny CDN), the handler must gzip-encode:

```go
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    gz := gzip.NewWriter(w)
    gz.Write([]byte("HIT|200|1507167062421|412|...\n"))
    gz.Close() // MUST close — flushes the compressor
}))
```

### Testability — configurable base URL

Code that hardcodes a URL can't be tested with a fake server. Make the base URL a field:

```go
// ❌ Hardcoded — can't redirect to test server
type Fetcher struct {
    apiKey string
}
url := fmt.Sprintf("https://logging.bunnycdn.com/%s", date)

// ✅ Configurable — pass server.URL in tests
type Fetcher struct {
    baseURL string // "https://logging.bunnycdn.com" in prod, server.URL in tests
    apiKey  string
}
url := fmt.Sprintf("%s/%s", f.baseURL, date)
```

## Naming Conventions

| Pattern | Meaning |
|---|---|
| `TestFunctionName` | Tests a specific function |
| `TestFunctionName_Scenario` | Tests one scenario of a function (underscore is visual separator, not special syntax) |
| `TestTypeName_MethodName` | Tests a method on a type |
| `testLogLine` (lowercase) | Test helper variable — not a test function, just shared data |

## Key Imports

```go
import (
    "testing"           // always needed
    "context"           // context.Background() for method calls
    "net/http/httptest" // fake HTTP server
    "compress/gzip"     // gzip encoding in test handlers
    "time"              // time.Date() for fixed test dates
)
```
