package http_handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/golang-jwt/jwt/v5"
	"strings"
	"sync"
	"time"
)

const (
	bearerPrefix = "Bearer "
)

type Middleware interface {
	Logger(c fiber.Ctx) error
	RequiredGMRoles(c fiber.Ctx) error
	RequiredAuthentication(c fiber.Ctx) error
	RevokeToken(jti string)
}

type middleware struct {
	jwtSecret     string
	revokedTokens sync.Map
}

func (m *middleware) RevokeToken(jti string) {
	if jti != "" {
		m.revokedTokens.Store(jti, time.Now())
	}
}

func (m *middleware) isRevoked(jti string) bool {
	if jti == "" {
		return false
	}
	_, ok := m.revokedTokens.Load(jti)
	return ok
}

func NewMiddleware(jwtSecret string) Middleware {
	return &middleware{
		jwtSecret: jwtSecret,
	}
}

func (m *middleware) Logger(c fiber.Ctx) error {
	return c.Next()
}

func NewRateLimiter(max int, window time.Duration) fiber.Handler {
	if max <= 0 {
		max = 120
	}
	if window <= 0 {
		window = time.Minute
	}
	return limiter.New(limiter.Config{Max: max, Expiration: window, KeyGenerator: func(c fiber.Ctx) string { return c.IP() }, LimitReached: func(c fiber.Ctx) error { return fiber.NewError(fiber.StatusTooManyRequests, "rate limit exceeded") }})
}

func (m *middleware) RequiredAuthentication(c fiber.Ctx) error {
	h := c.Get("Authorization")
	if !strings.HasPrefix(h, bearerPrefix) {
		return fiber.NewError(fiber.StatusUnauthorized, "missing bearer token")
	}
	raw := strings.TrimPrefix(h, bearerPrefix)
	token, err := jwt.Parse(raw, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}
		return []byte(m.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if jti, _ := claims["jti"].(string); m.isRevoked(jti) {
			return fiber.NewError(fiber.StatusUnauthorized, "token revoked")
		}
		c.Locals("sub", claims["sub"])
		c.Locals("role", claims["role"])
		c.Locals("jti", claims["jti"])
	}
	return c.Next()

}

func (m *middleware) RequiredGMRoles(c fiber.Ctx) error {
	h := c.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return fiber.NewError(401, "missing bearer token")
	}
	raw := strings.TrimPrefix(h, "Bearer ")
	token, err := jwt.Parse(raw, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fiber.NewError(401, "invalid token")
		}
		return []byte(m.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return fiber.NewError(401, "invalid token")
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if jti, _ := claims["jti"].(string); m.isRevoked(jti) {
			return fiber.NewError(401, "token revoked")
		}
		c.Locals("gm", claims["sub"])
		c.Locals("role", claims["role"])
		c.Locals("jti", claims["jti"])
	}
	return c.Next()
}
