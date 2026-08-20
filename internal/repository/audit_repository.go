package repository

import (
	"context"
	"database/sql"
)

type AuditLog struct {
	ID                                                     int64
	RequestID, GMUsername, Action, Target, Reason, Outcome string
	CreatedAt                                              any
}
type AuditRepository interface {
	List(ctx context.Context, limit uint) ([]AuditLog, error)
}
type auditRepository struct{ db *sql.DB }

func NewAuditRepository(db *sql.DB) AuditRepository { return &auditRepository{db: db} }
func (r *auditRepository) List(ctx context.Context, limit uint) ([]AuditLog, error) {
	if limit == 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	rows, err := r.db.QueryContext(ctx, "select top (?) id,request_id,gm_username,action_name,target_name,reason,outcome,created_at from gm_audit_logs order by id desc", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	logs := make([]AuditLog, 0)
	for rows.Next() {
		var l AuditLog
		var target, reason sql.NullString
		if err := rows.Scan(&l.ID, &l.RequestID, &l.GMUsername, &l.Action, &target, &reason, &l.Outcome, &l.CreatedAt); err != nil {
			return nil, err
		}
		l.Target = target.String
		l.Reason = reason.String
		logs = append(logs, l)
	}
	return logs, rows.Err()
}
