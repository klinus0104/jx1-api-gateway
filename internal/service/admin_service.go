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
	"github.com/klinus0104/jx1-api-gateway/pkg/heaven"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidAdminCredentials = errors.New("invalid admin credentials")

type AdminService interface {
	Login(ctx context.Context, req AdminLoginRequest) (dto.AdminLoginDTO, error)
	Me(ctx context.Context, username string) (dto.AdminMeDTO, error)
	ListAccounts(ctx context.Context, req ListAccountsRequest) (dto.AccountListDTO, error)
	GetAccount(ctx context.Context, req AccountRequest) (dto.AccountDTO, error)
	BlockAccount(ctx context.Context, req AccountActionRequest) (dto.AccountActionDTO, error)
	UnblockAccount(ctx context.Context, req AccountActionRequest) (dto.AccountActionDTO, error)
	ResetPassword(ctx context.Context, req ResetPasswordRequest) (dto.AccountActionDTO, error)
	KickPlayer(ctx context.Context, req AccountActionRequest) (dto.AccountActionDTO, error)
	GetSessions(ctx context.Context, req AccountRequest) (dto.SessionDTO, error)
	ListAuditLogs(ctx context.Context, req AuditLogRequest) (dto.AuditLogListDTO, error)
}
type adminService struct {
	accountRepo  repository.AccountRepository
	auditRepo    repository.AuditRepository
	gmRepo       repository.GMUserRepository
	heavenClient *heaven.Client
	jwtSecret    string
	tokenTTL     time.Duration
}
type AdminLoginRequest struct {
	Username string
	Password string
}
type ListAccountsRequest struct {
	Q     string
	Limit uint
	Page  uint
}
type AccountRequest struct{ Account string }
type AccountActionRequest struct {
	Account  string
	Reason   string
	TicketID string
}
type ResetPasswordRequest struct {
	Account  string
	Password string
	Reason   string
	TicketID string
}
type AuditLogRequest struct{ Limit uint }

var ErrInvalidAdminRequest = errors.New("invalid admin request")

func (a *adminService) Login(ctx context.Context, req AdminLoginRequest) (dto.AdminLoginDTO, error) {
	if a.gmRepo == nil || strings.TrimSpace(req.Username) == "" || req.Password == "" || a.jwtSecret == "" {
		return dto.AdminLoginDTO{}, ErrInvalidAdminCredentials
	}
	user, role, err := a.gmRepo.FindActive(ctx, strings.TrimSpace(req.Username))
	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return dto.AdminLoginDTO{}, ErrInvalidAdminCredentials
	}
	if role == "" {
		role = authz.RoleViewer
	}
	if a.tokenTTL <= 0 {
		a.tokenTTL = 24 * time.Hour
	}
	expiresAt := time.Now().Add(a.tokenTTL)
	jtiBytes := make([]byte, 16)
	if _, err = rand.Read(jtiBytes); err != nil {
		return dto.AdminLoginDTO{}, err
	}
	claims := jwt.MapClaims{"sub": user.Username, "role": role, "jti": hex.EncodeToString(jtiBytes), "exp": expiresAt.Unix()}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(a.jwtSecret))
	if err != nil {
		return dto.AdminLoginDTO{}, err
	}
	return dto.AdminLoginDTO{Token: token, ExpiresIn: int(a.tokenTTL.Seconds()), Role: role}, nil
}

func (a *adminService) Me(ctx context.Context, username string) (dto.AdminMeDTO, error) {
	username = strings.TrimSpace(username)
	if a.gmRepo == nil || username == "" {
		return dto.AdminMeDTO{}, ErrInvalidAdminRequest
	}
	user, role, err := a.gmRepo.FindActive(ctx, username)
	if err != nil || user == nil {
		return dto.AdminMeDTO{}, ErrInvalidAdminCredentials
	}
	if role == "" {
		role = authz.RoleViewer
	}
	return dto.AdminMeDTO{Username: user.Username, Role: role, Active: user.IsActive}, nil
}

func (a *adminService) ListAccounts(ctx context.Context, req ListAccountsRequest) (dto.AccountListDTO, error) {
	page, err := a.accountRepo.GetListAccounts(ctx, req.Q, req.Page, req.Limit)
	if err != nil {
		return dto.AccountListDTO{}, err
	}
	items := make([]dto.AccountDTO, 0, len(page.Items))
	for _, account := range page.Items {
		if account == nil {
			continue
		}
		item := dto.AccountDTO{Name: account.CAccName, Banned: account.BIsBanned, Online: account.IClientID.Valid && account.IClientID.Int64 != 0}
		if account.IClientID.Valid {
			item.ClientID = account.IClientID.Int64
		}
		if account.NUserIP.Valid {
			item.UserIP = int64(account.NUserIP.Int)
		}
		items = append(items, item)
	}
	return dto.AccountListDTO{Items: items, Page: page.Page, PageSize: page.PageSize, Total: page.Total, TotalPages: page.TotalPages}, nil
}

func (a *adminService) GetAccount(ctx context.Context, req AccountRequest) (dto.AccountDTO, error) {
	if strings.TrimSpace(req.Account) == "" {
		return dto.AccountDTO{}, ErrInvalidAdminRequest
	}
	account, err := a.accountRepo.GetAccount(ctx, strings.TrimSpace(req.Account))
	if err != nil {
		return dto.AccountDTO{}, err
	}
	if account == nil {
		return dto.AccountDTO{}, ErrInvalidAdminRequest
	}
	item := dto.AccountDTO{Name: account.CAccName, Banned: account.BIsBanned, Online: account.IClientID.Valid && account.IClientID.Int64 != 0}
	if account.IClientID.Valid {
		item.ClientID = account.IClientID.Int64
	}
	if account.NUserIP.Valid {
		item.UserIP = int64(account.NUserIP.Int)
	}
	return item, nil
}

func (a *adminService) BlockAccount(ctx context.Context, req AccountActionRequest) (dto.AccountActionDTO, error) {
	name := strings.TrimSpace(req.Account)
	if name == "" || strings.TrimSpace(req.Reason) == "" {
		return dto.AccountActionDTO{}, ErrInvalidAdminRequest
	}
	// Ban and disconnect are one administrative operation: terminate the
	// active realtime session before persisting the banned state.
	if a.heavenClient != nil {
		if err := a.heavenClient.LegacyKick(ctx, name); err != nil {
			return dto.AccountActionDTO{}, err
		}
	}
	if err := a.accountRepo.Block(ctx, name); err != nil {
		return dto.AccountActionDTO{}, err
	}
	return dto.AccountActionDTO{Success: true, Account: name, Action: authz.ActionBlock}, nil
}
func (a *adminService) UnblockAccount(ctx context.Context, req AccountActionRequest) (dto.AccountActionDTO, error) {
	return a.mutateAccount(ctx, req, authz.ActionUnblock, a.accountRepo.Unblock)
}
func (a *adminService) KickPlayer(ctx context.Context, req AccountActionRequest) (dto.AccountActionDTO, error) {
	if a.heavenClient != nil {
		if err := a.heavenClient.LegacyKick(ctx, strings.TrimSpace(req.Account)); err != nil {
			return dto.AccountActionDTO{}, err
		}
	}
	return a.mutateAccount(ctx, req, authz.ActionKick, a.accountRepo.Kick)
}
func (a *adminService) mutateAccount(ctx context.Context, req AccountActionRequest, action string, fn func(context.Context, string) error) (dto.AccountActionDTO, error) {
	name := strings.TrimSpace(req.Account)
	if name == "" || strings.TrimSpace(req.Reason) == "" {
		return dto.AccountActionDTO{}, ErrInvalidAdminRequest
	}
	if err := fn(ctx, name); err != nil {
		return dto.AccountActionDTO{}, err
	}
	return dto.AccountActionDTO{Success: true, Account: name, Action: action}, nil
}
func (a *adminService) ResetPassword(ctx context.Context, req ResetPasswordRequest) (dto.AccountActionDTO, error) {
	if strings.TrimSpace(req.Account) == "" || len(req.Password) < 6 || strings.TrimSpace(req.Reason) == "" {
		return dto.AccountActionDTO{}, ErrInvalidAdminRequest
	}
	if err := a.accountRepo.ResetPassword(ctx, strings.TrimSpace(req.Account), req.Password); err != nil {
		return dto.AccountActionDTO{}, err
	}
	return dto.AccountActionDTO{Success: true, Account: strings.TrimSpace(req.Account), Action: authz.ActionResetPassword}, nil
}

func (a *adminService) GetSessions(ctx context.Context, req AccountRequest) (dto.SessionDTO, error) {
	name := strings.TrimSpace(req.Account)
	if name == "" {
		return dto.SessionDTO{}, ErrInvalidAdminRequest
	}
	account, err := a.accountRepo.GetAccount(ctx, name)
	if err != nil {
		return dto.SessionDTO{}, err
	}
	if account == nil {
		return dto.SessionDTO{}, ErrInvalidAdminRequest
	}
	s := dto.SessionDTO{Account: account.CAccName, Online: account.IClientID.Valid && account.IClientID.Int64 != 0}
	if account.IClientID.Valid {
		s.ClientID = account.IClientID.Int64
	}
	if account.NUserIP.Valid {
		s.UserIP = int64(account.NUserIP.Int)
	}
	return s, nil
}

func (a *adminService) ListAuditLogs(ctx context.Context, req AuditLogRequest) (dto.AuditLogListDTO, error) {
	if a.auditRepo == nil {
		return dto.AuditLogListDTO{}, errors.New("audit repository unavailable")
	}
	logs, err := a.auditRepo.List(ctx, req.Limit)
	if err != nil {
		return dto.AuditLogListDTO{}, err
	}
	items := make([]dto.AuditLogDTO, 0, len(logs))
	for _, l := range logs {
		items = append(items, dto.AuditLogDTO{ID: l.ID, RequestID: l.RequestID, GMUsername: l.GMUsername, Action: l.Action, Target: l.Target, Reason: l.Reason, Outcome: l.Outcome, CreatedAt: l.CreatedAt})
	}
	return dto.AuditLogListDTO{Items: items}, nil
}

func NewAdminService(accountRepo repository.AccountRepository, gmRepo repository.GMUserRepository, auditRepo repository.AuditRepository, heavenClient *heaven.Client, jwtSecret string, tokenTTL time.Duration) AdminService {
	return &adminService{accountRepo: accountRepo, gmRepo: gmRepo, auditRepo: auditRepo, heavenClient: heavenClient, jwtSecret: jwtSecret, tokenTTL: tokenTTL}
}
