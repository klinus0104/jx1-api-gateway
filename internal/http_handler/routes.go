package http_handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/klinus0104/jx1-api-gateway/docs"
	"github.com/klinus0104/jx1-api-gateway/internal/service"
	"time"
)

type Server interface {
	App() *fiber.App
	Listen() error
	Shutdown() error
}

type server struct {
	app        *fiber.App
	address    string
	middleware Middleware
}

func NewServer(
	address string,
	middleware Middleware,
	adminService service.AdminService,
	playerService service.PlayerService,
	authService service.AuthService,
) Server {
	adminHandler := NewAdminHandler(adminService)
	playerHandler := NewPlayerHandler(playerService)
	authHandler := NewAuthHandler(authService)
	app := fiber.New(fiber.Config{
		AppName:      "jx-api-gateway",
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	})
	app.Get("/healthz", func(c fiber.Ctx) error { return c.JSON(map[string]any{"status": "ok"}) })
	app.Get("/openapi.yaml", func(c fiber.Ctx) error { c.Type("yaml"); return c.SendString(docs.OpenAPISpec) })
	app.Get("/docs", func(c fiber.Ctx) error { return c.Type("html").SendString(docs.SwaggerHTML) })

	// Middlewares global
	app.Use(recover.New())
	app.Use(middleware.Logger)
	app.Use(NewRateLimiter(120, time.Minute))
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	api := app.Group("/api/v1")
	playerAPI := api.Group("/player")
	playerAPI.Post("/accounts/register", playerHandler.Register)
	playerAPI.Post("/auth/login", playerHandler.Login)
	protected := api.Group("", middleware.RequiredAuthentication)
	protected.Post("/auth/logout", authHandler.Logout)

	protected.Get("/me", adminHandler.Me)

	api.Post("/admin/auth/login", adminHandler.Login)
	admin := protected.Group("/admin", middleware.RequiredGMRoles)
	admin.Get("/accounts", adminHandler.ListAccounts)
	admin.Get("/accounts/:name", adminHandler.GetAccount)
	admin.Get("/accounts/:name/sessions", adminHandler.Sessions)
	admin.Get("/audit-logs", adminHandler.AuditLogs)
	admin.Post("/accounts/:name/block", adminHandler.Block)
	admin.Post("/accounts/:name/unblock", adminHandler.Unblock)
	admin.Post("/accounts/:name/reset-password", adminHandler.ResetPassword)
	admin.Post("/players/:id/kick", adminHandler.Kick)

	player := protected.Group("/player")
	player.Get("/profile", playerHandler.Profile)
	player.Post("/auth/logout", func(c fiber.Ctx) error { return c.JSON(map[string]any{"success": true}) })
	player.Post("/accounts/change-password", playerHandler.ChangePassword)
	//_ = log

	return &server{
		address: address,
		app:     app,
	}
}

func (s *server) App() *fiber.App {
	return s.app
}

func (s *server) Listen() error {
	return s.app.Listen(s.address)
}

func (s *server) Shutdown() error {
	return s.app.Shutdown()
}
