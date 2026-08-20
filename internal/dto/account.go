package dto

type AccountDTO struct {
	Name     string `json:"name"`
	ClientID int64  `json:"client_id"`
	UserIP   int64  `json:"user_ip"`
	Online   bool   `json:"online"`
	Banned   bool   `json:"banned"`
}

type AccountListDTO struct {
	Items      []AccountDTO `json:"items"`
	Page       uint         `json:"page"`
	PageSize   uint         `json:"page_size"`
	Total      uint         `json:"total"`
	TotalPages uint         `json:"total_pages"`
}

type AdminLoginDTO struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
	Role      string `json:"role"`
}

type AdminMeDTO struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	Active   bool   `json:"active"`
}

type AccountActionDTO struct {
	Success bool   `json:"success"`
	Account string `json:"account"`
	Action  string `json:"action"`
}

type PlayerLoginDTO struct {
	Token     string `json:"token"`
	Account   string `json:"account"`
	ExpiresIn int    `json:"expires_in"`
}
type PlayerRegisterDTO struct {
	Success bool   `json:"success"`
	Account string `json:"account"`
}
type PlayerProfileDTO struct {
	Account string `json:"account"`
	Role    string `json:"role"`
}

type SessionDTO struct {
	Account  string `json:"account"`
	ClientID int64  `json:"client_id"`
	UserIP   int64  `json:"user_ip"`
	Online   bool   `json:"online"`
}
type AuditLogDTO struct {
	ID         int64  `json:"id"`
	RequestID  string `json:"request_id"`
	GMUsername string `json:"gm_username"`
	Action     string `json:"action"`
	Target     string `json:"target,omitempty"`
	Reason     string `json:"reason,omitempty"`
	Outcome    string `json:"outcome"`
	CreatedAt  any    `json:"created_at"`
}
type AuditLogListDTO struct {
	Items []AuditLogDTO `json:"items"`
}
