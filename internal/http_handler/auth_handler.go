package http_handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/klinus0104/jx1-api-gateway/internal/service"
)

// AuthHandler adapts authentication-related HTTP requests to AuthService.
type AuthHandler struct{ service service.AuthService }

func NewAuthHandler(s service.AuthService) *AuthHandler { return &AuthHandler{service: s} }

func (h *AuthHandler) Logout(c fiber.Ctx) error {
	jti, _ := c.Locals("jti").(string)
	if err := h.service.Logout(jti); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid logout token")
	}
	return c.JSON(map[string]any{"success": true})
}
