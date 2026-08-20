package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/klinus0104/jx1-api-gateway/internal/authz"
	"github.com/klinus0104/jx1-api-gateway/internal/dto"
	"github.com/klinus0104/jx1-api-gateway/internal/repository"
)

var ErrInvalidPlayerRequest = errors.New("invalid player request")
var ErrInvalidPlayerCredentials = errors.New("invalid player credentials")

type PlayerService interface {
	Register(ctx context.Context, req PlayerRegisterRequest) (dto.PlayerRegisterDTO, error)
	Login(ctx context.Context, req PlayerLoginRequest) (dto.PlayerLoginDTO, error)
	Profile(ctx context.Context, account string) (dto.PlayerProfileDTO, error)
	ChangePassword(ctx context.Context, req PlayerChangePasswordRequest) error
}

type playerService struct {
	accountRepo repository.AccountRepository
	jwtSecret   string
	tokenTTL    time.Duration
}
type PlayerRegisterRequest struct {
	Account  string
	Password string
}
type PlayerLoginRequest struct {
	Account  string
	Password string
}
type PlayerChangePasswordRequest struct {
	Account         string
	CurrentPassword string
	NewPassword     string
}

func (p *playerService) Register(ctx context.Context, req PlayerRegisterRequest) (dto.PlayerRegisterDTO, error) {
	account := strings.TrimSpace(req.Account)
	if len(account) < 6 || len(req.Password) < 6 {
		return dto.PlayerRegisterDTO{}, ErrInvalidPlayerRequest
	}
	if err := p.accountRepo.CreateAccount(ctx, account, req.Password); err != nil {
		return dto.PlayerRegisterDTO{}, err
	}
	return dto.PlayerRegisterDTO{Success: true, Account: account}, nil
}

func (p *playerService) Login(ctx context.Context, req PlayerLoginRequest) (dto.PlayerLoginDTO, error) {
	account := strings.TrimSpace(req.Account)
	if account == "" || req.Password == "" || p.jwtSecret == "" {
		return dto.PlayerLoginDTO{}, ErrInvalidPlayerCredentials
	}
	ok, err := p.accountRepo.VerifyPlayer(ctx, account, req.Password)
	if err != nil || !ok {
		return dto.PlayerLoginDTO{}, ErrInvalidPlayerCredentials
	}
	if p.tokenTTL <= 0 {
		p.tokenTTL = 24 * time.Hour
	}
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return dto.PlayerLoginDTO{}, err
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": account, "role": authz.RolePlayer, "jti": hex.EncodeToString(b), "exp": time.Now().Add(p.tokenTTL).Unix()}).SignedString([]byte(p.jwtSecret))
	if err != nil {
		return dto.PlayerLoginDTO{}, err
	}
	return dto.PlayerLoginDTO{Token: token, Account: account, ExpiresIn: int(p.tokenTTL.Seconds())}, nil
}

func (p *playerService) Profile(ctx context.Context, account string) (dto.PlayerProfileDTO, error) {
	account = strings.TrimSpace(account)
	if account == "" {
		return dto.PlayerProfileDTO{}, ErrInvalidPlayerRequest
	}
	if _, err := p.accountRepo.GetAccount(ctx, account); err != nil {
		return dto.PlayerProfileDTO{}, err
	}
	return dto.PlayerProfileDTO{Account: account, Role: authz.RolePlayer}, nil
}

func (p *playerService) ChangePassword(ctx context.Context, req PlayerChangePasswordRequest) error {
	if strings.TrimSpace(req.Account) == "" || req.CurrentPassword == "" || len(req.NewPassword) < 6 {
		return ErrInvalidPlayerRequest
	}
	return p.accountRepo.ChangePassword(ctx, strings.TrimSpace(req.Account), req.CurrentPassword, req.NewPassword)
}

func NewPlayerService(accountRepo repository.AccountRepository, jwtSecret string, tokenTTL time.Duration) PlayerService {
	return &playerService{accountRepo: accountRepo, jwtSecret: jwtSecret, tokenTTL: tokenTTL}
}
