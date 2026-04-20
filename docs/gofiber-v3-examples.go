package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
)

// Example 1: Basic Server Setup
func exampleBasicSetup() {
	app := fiber.New()

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	log.Fatal(app.Listen(":3000"))
}

// Example 2: Route Parameters and Query Strings
func exampleRouting(app *fiber.App) {
	// Route with parameters
	app.Get("/users/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		return c.SendString("User ID: " + id)
	})

	// Route with nested parameters
	app.Get("/users/:userId/posts/:postId", func(c fiber.Ctx) error {
		userId := c.Params("userId")
		postId := c.Params("postId")
		return c.JSON(fiber.Map{
			"userId": userId,
			"postId": postId,
		})
	})

	// Query parameters
	app.Get("/search", func(c fiber.Ctx) error {
		query := c.Query("q", "")
		page := c.QueryInt("page", 1)
		return c.JSON(fiber.Map{
			"search": query,
			"page":   page,
		})
	})
}

// Example 3: Route Grouping with Middleware
func exampleRouteGrouping(app *fiber.App) {
	// Base API group
	api := app.Group("/api")

	// Version 1 group with middleware
	v1 := api.Group("/v1", func(c fiber.Ctx) error {
		c.Set("X-API-Version", "1.0")
		c.Set("X-Timestamp", time.Now().String())
		return c.Next()
	})

	v1.Get("/status", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Version 2 group
	v2 := api.Group("/v2", func(c fiber.Ctx) error {
		c.Set("X-API-Version", "2.0")
		return c.Next()
	})

	v2.Get("/status", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"version": "2.0",
		})
	})
}

// Example 4: Global and Scoped Middleware
func exampleMiddleware(app *fiber.App) {
	// Global middleware - runs on all requests
	app.Use(func(c fiber.Ctx) error {
		log.Printf("[%s] %s %s", c.Method(), c.Path(), c.IP())
		return c.Next()
	})

	// Middleware for specific prefix
	app.Use("/api", func(c fiber.Ctx) error {
		c.Set("X-API-Request", "true")
		return c.Next()
	})

	// Multiple middleware on same route
	app.Get("/protected",
		authenticateMiddleware(),
		authorizationMiddleware(),
		func(c fiber.Ctx) error {
			return c.JSON(fiber.Map{"message": "Access granted"})
		},
	)
}

// Middleware functions
func authenticateMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		token := c.Get("Authorization")
		if token == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "Missing token")
		}
		c.Locals("user", "user123")
		return c.Next()
	}
}

func authorizationMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		user := c.Locals("user")
		if user == nil {
			return fiber.NewError(fiber.StatusForbidden, "Unauthorized")
		}
		return c.Next()
	}
}

// Example 5: Request Body Parsing
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func exampleRequestParsing(app *fiber.App) {
	// JSON binding
	app.Post("/users", func(c fiber.Ctx) error {
		user := new(User)

		if err := c.Bind().Body(user); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON")
		}

		return c.Status(fiber.StatusCreated).JSON(user)
	})

	// JSON-specific binding
	app.Post("/users/json-only", func(c fiber.Ctx) error {
		user := new(User)

		if err := c.Bind().JSON(user); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON")
		}

		return c.Status(fiber.StatusCreated).JSON(user)
	})

	// Get with default values
	app.Get("/users/search", func(c fiber.Ctx) error {
		name := c.Query("name", "")
		sort := c.Query("sort", "id")
		limit := c.QueryInt("limit", 10)

		return c.JSON(fiber.Map{
			"name":  name,
			"sort":  sort,
			"limit": limit,
		})
	})
}

// Example 6: Response Handling
func exampleResponses(app *fiber.App) {
	// Send string
	app.Get("/text", func(c fiber.Ctx) error {
		return c.SendString("Plain text response")
	})

	// Send JSON
	app.Get("/json", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello",
			"status":  "success",
		})
	})

	// Send with status code
	app.Post("/create", func(c fiber.Ctx) error {
		user := User{ID: 1, Name: "John", Email: "john@example.com"}
		return c.Status(fiber.StatusCreated).JSON(user)
	})

	// Send file
	app.Get("/download", func(c fiber.Ctx) error {
		return c.SendFile("./files/document.pdf")
	})

	// Custom headers
	app.Get("/custom-headers", func(c fiber.Ctx) error {
		c.Set("X-Custom-Header", "CustomValue")
		c.Set("Content-Type", "application/json")
		return c.SendString(`{"status":"ok"}`)
	})

	// Set cookies
	app.Get("/set-cookie", func(c fiber.Ctx) error {
		c.Cookie(&fiber.Cookie{
			Name:     "session_id",
			Value:    "abc123xyz",
			Expires:  time.Now().Add(24 * time.Hour),
			HTTPOnly: true,
			Secure:   true,
			SameSite: "Lax",
		})
		return c.SendString("Cookie set")
	})
}

// Example 7: Error Handling
func exampleErrorHandling(app *fiber.App) {
	// Simple error return
	app.Get("/divide/:a/:b", func(c fiber.Ctx) error {
		a, errA := parseIntParam(c, "a")
		b, errB := parseIntParam(c, "b")

		if errA != nil || errB != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid parameters")
		}

		if b == 0 {
			return fiber.NewError(fiber.StatusBadRequest, "Division by zero")
		}

		return c.JSON(fiber.Map{
			"result": a / b,
		})
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

		// Log the error
		log.Printf("Error [%d]: %s", code, message)

		// Return JSON error response
		return c.Status(code).JSON(fiber.Map{
			"error":  message,
			"status": code,
			"path":   c.Path(),
			"method": c.Method(),
		})
	})
}

func parseIntParam(c fiber.Ctx, key string) (int, error) {
	val := c.Params(key)
	if val == "" {
		return 0, errors.New("missing parameter")
	}
	var result int
	_, err := fmt.Sscanf(val, "%d", &result)
	return result, err
}

// Example 8: Async Operations with Context
func exampleAsyncOperations(app *fiber.App) {
	app.Get("/async-task", func(c fiber.Ctx) error {
		ctx := c.UserContext()

		resultChan := make(chan string)
		go performBackgroundTask(ctx, resultChan)

		select {
		case result := <-resultChan:
			return c.JSON(fiber.Map{"result": result})
		case <-ctx.Done():
			return fiber.NewError(fiber.StatusRequestTimeout, "Context cancelled")
		}
	})

	app.Get("/timeout-task", func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		resultChan := make(chan string)
		go performSlowTask(resultChan)

		select {
		case result := <-resultChan:
			return c.JSON(fiber.Map{"result": result})
		case <-ctx.Done():
			return fiber.NewError(fiber.StatusRequestTimeout, "Operation timed out")
		}
	})
}

func performBackgroundTask(ctx context.Context, result chan<- string) {
	select {
	case <-time.After(1 * time.Second):
		result <- "task completed"
	case <-ctx.Done():
		// Task was cancelled
	}
}

func performSlowTask(result chan<- string) {
	time.Sleep(3 * time.Second)
	result <- "slow task done"
}

// Example 9: CRUD API
func exampleCRUDAPI(app *fiber.App) {
	// In-memory storage
	var users []User
	nextID := 1

	api := app.Group("/api/users")

	// Create
	api.Post("/", func(c fiber.Ctx) error {
		user := new(User)
		if err := c.Bind().Body(user); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON")
		}

		user.ID = nextID
		nextID++
		users = append(users, *user)

		return c.Status(fiber.StatusCreated).JSON(user)
	})

	// Read all
	api.Get("/", func(c fiber.Ctx) error {
		return c.JSON(users)
	})

	// Read one
	api.Get("/:id", func(c fiber.Ctx) error {
		id, _ := strconv.Atoi(c.Params("id"))

		for _, user := range users {
			if user.ID == id {
				return c.JSON(user)
			}
		}

		return fiber.NewError(fiber.StatusNotFound, "User not found")
	})

	// Update
	api.Put("/:id", func(c fiber.Ctx) error {
		id, _ := strconv.Atoi(c.Params("id"))

		updatedUser := new(User)
		if err := c.Bind().Body(updatedUser); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid JSON")
		}

		for i, user := range users {
			if user.ID == id {
				updatedUser.ID = id
				users[i] = *updatedUser
				return c.JSON(updatedUser)
			}
		}

		return fiber.NewError(fiber.StatusNotFound, "User not found")
	})

	// Delete
	api.Delete("/:id", func(c fiber.Ctx) error {
		id, _ := strconv.Atoi(c.Params("id"))

		for i, user := range users {
			if user.ID == id {
				users = append(users[:i], users[i+1:]...)
				return c.SendStatus(fiber.StatusNoContent)
			}
		}

		return fiber.NewError(fiber.StatusNotFound, "User not found")
	})
}

// Example 10: Request Headers and Cookies
func exampleHeadersAndCookies(app *fiber.App) {
	app.Get("/headers", func(c fiber.Ctx) error {
		authorization := c.Get("Authorization")
		contentType := c.Get("Content-Type")
		userAgent := c.Get("User-Agent")

		return c.JSON(fiber.Map{
			"authorization": authorization,
			"content_type":  contentType,
			"user_agent":    userAgent,
		})
	})

	app.Get("/cookies", func(c fiber.Ctx) error {
		sessionID := c.Cookies("session_id", "no-session")
		token := c.Cookies("token", "no-token")

		return c.JSON(fiber.Map{
			"session_id": sessionID,
			"token":      token,
		})
	})

	app.Get("/client-info", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"ip":       c.IP(),
			"hostname": c.Hostname(),
			"method":   c.Method(),
			"path":     c.Path(),
		})
	})
}
