package model

type AccountSnapshot struct {
	Name     string
	ClientID int64
	UserIP   int64
	Online   bool
}
type SessionSnapshot struct {
	Account  string
	ClientID int64
	UserIP   int64
	Online   bool
}
