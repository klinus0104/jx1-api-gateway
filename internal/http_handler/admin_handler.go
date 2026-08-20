package http_handler

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/klinus0104/jx1-api-gateway/internal/service"
)

type AdminHandler struct{ service service.AdminService }

func NewAdminHandler(s service.AdminService) *AdminHandler { return &AdminHandler{service: s} }
func (h *AdminHandler) Login(c fiber.Ctx) error {
	var body service.AdminLoginRequest
	if err := c.Bind().Body(&body); err != nil {
		return fiber.NewError(400, "invalid request")
	}
	out, err := h.service.Login(c.Context(), body)
	if err != nil {
		return fiber.NewError(401, "invalid credentials")
	}
	return c.JSON(out)
}
func (h *AdminHandler) Me(c fiber.Ctx) error {
	user, _ := c.Locals("sub").(string)
	out, err := h.service.Me(c.Context(), user)
	if err != nil {
		return fiber.NewError(404, "admin user not found")
	}
	return c.JSON(out)
}
func (h *AdminHandler) ListAccounts(c fiber.Ctx) error {
	out, err := h.service.ListAccounts(c.Context(), service.ListAccountsRequest{Q: strings.TrimSpace(c.Query("q")), Page: queryUint(c.Query("page")), Limit: queryUint(c.Query("limit"))})
	if err != nil {
		return fiber.NewError(500, "account query failed")
	}
	return c.JSON(out)
}
func (h *AdminHandler) GetAccount(c fiber.Ctx) error {
	out, err := h.service.GetAccount(c.Context(), service.AccountRequest{Account: c.Params("name")})
	if err != nil {
		return fiber.NewError(404, "account not found")
	}
	return c.JSON(out)
}
func (h *AdminHandler) Sessions(c fiber.Ctx) error {
	out, err := h.service.GetSessions(c.Context(), service.AccountRequest{Account: c.Params("name")})
	if err != nil {
		return fiber.NewError(404, "account not found")
	}
	return c.JSON(out)
}
func (h *AdminHandler) AuditLogs(c fiber.Ctx) error {
	out, err := h.service.ListAuditLogs(c.Context(), service.AuditLogRequest{Limit: queryUint(c.Query("limit"))})
	if err != nil {
		return fiber.NewError(500, "audit query failed")
	}
	return c.JSON(out)
}

func queryUint(value string) uint {
	n, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0
	}
	return uint(n)
}
func (h *AdminHandler) Block(c fiber.Ctx) error {
	return h.action(c, func(r service.AccountActionRequest) (any, error) { return h.service.BlockAccount(c.Context(), r) })
}
func (h *AdminHandler) Unblock(c fiber.Ctx) error {
	return h.action(c, func(r service.AccountActionRequest) (any, error) { return h.service.UnblockAccount(c.Context(), r) })
}
func (h *AdminHandler) Kick(c fiber.Ctx) error {
	return h.action(c, func(r service.AccountActionRequest) (any, error) { return h.service.KickPlayer(c.Context(), r) })
}
func (h *AdminHandler) ResetPassword(c fiber.Ctx) error {
	var b struct{ Password, Reason, TicketID string }
	if err := c.Bind().Body(&b); err != nil {
		return fiber.NewError(400, "invalid request")
	}
	out, err := h.service.ResetPassword(c.Context(), service.ResetPasswordRequest{Account: c.Params("name"), Password: b.Password, Reason: b.Reason, TicketID: b.TicketID})
	if err != nil {
		return fiber.NewError(400, "invalid password reset request")
	}
	return c.JSON(out)
}
func (h *AdminHandler) action(c fiber.Ctx, call func(service.AccountActionRequest) (any, error)) error {
	var b struct{ Reason, TicketID string }
	if err := c.Bind().Body(&b); err != nil {
		return fiber.NewError(400, "invalid request")
	}
	out, err := call(service.AccountActionRequest{Account: c.Params("name"), Reason: b.Reason, TicketID: b.TicketID})
	if err != nil {
		if errors.Is(err, service.ErrInvalidAdminRequest) {
			return fiber.NewError(400, "account and reason are required")
		}
		return fiber.NewError(500, "account action failed")
	}
	return c.JSON(out)
}
