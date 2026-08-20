package http_handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/klinus0104/jx1-api-gateway/internal/service"
)

type PlayerHandler struct{ service service.PlayerService }

func NewPlayerHandler(s service.PlayerService) *PlayerHandler { return &PlayerHandler{service: s} }
func (h *PlayerHandler) Register(c fiber.Ctx) error {
	var b service.PlayerRegisterRequest
	if err := c.Bind().Body(&b); err != nil {
		return fiber.NewError(400, "invalid request")
	}
	out, err := h.service.Register(c.Context(), b)
	if err != nil {
		return fiber.NewError(400, "account and password must be at least 6 characters")
	}
	return c.Status(201).JSON(out)
}
func (h *PlayerHandler) Login(c fiber.Ctx) error {
	var b service.PlayerLoginRequest
	if err := c.Bind().Body(&b); err != nil {
		return fiber.NewError(400, "invalid request")
	}
	out, err := h.service.Login(c.Context(), b)
	if err != nil {
		return fiber.NewError(401, "invalid player credentials")
	}
	return c.JSON(out)
}
func (h *PlayerHandler) Profile(c fiber.Ctx) error {
	account, _ := c.Locals("sub").(string)
	out, err := h.service.Profile(c.Context(), account)
	if err != nil {
		return fiber.NewError(404, "player not found")
	}
	return c.JSON(out)
}
func (h *PlayerHandler) ChangePassword(c fiber.Ctx) error {
	account, _ := c.Locals("sub").(string)
	var b struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := c.Bind().Body(&b); err != nil {
		return fiber.NewError(400, "invalid request")
	}
	if err := h.service.ChangePassword(c.Context(), service.PlayerChangePasswordRequest{Account: account, CurrentPassword: b.CurrentPassword, NewPassword: b.NewPassword}); err != nil {
		return fiber.NewError(401, "invalid current password")
	}
	return c.JSON(map[string]any{"success": true, "account": account})
}
