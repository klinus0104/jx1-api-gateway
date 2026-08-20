package service

import "errors"

var ErrInvalidLogoutToken = errors.New("invalid logout token")

type TokenRevoker interface{ RevokeToken(jti string) }

type AuthService interface{ Logout(jti string) error }

type authService struct{ revoker TokenRevoker }

func (a *authService) Logout(jti string) error {
	if jti == "" || a.revoker == nil {
		return ErrInvalidLogoutToken
	}
	a.revoker.RevokeToken(jti)
	return nil
}

func NewAuthService(revoker TokenRevoker) AuthService { return &authService{revoker: revoker} }
