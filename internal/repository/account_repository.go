package repository

import (
	"context"
	"database/sql"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/klinus0104/jx1-api-gateway/internal/db/models"
)

type AccountRepository interface {
	GetAccount(ctx context.Context, account string) (*models.AccountInfo, error)
	CreateAccount(ctx context.Context, account, password string) error
	VerifyPlayer(ctx context.Context, account, password string) (bool, error)
	ChangePassword(ctx context.Context, account, current, next string) error
	ResetPassword(ctx context.Context, account, password string) error
	Block(ctx context.Context, account string) error
	Unblock(ctx context.Context, account string) error
	Kick(ctx context.Context, account string) error
	GetListAccounts(ctx context.Context, q string, page uint, limit uint) (AccountPage, error)
}

type AccountPage struct {
	Items      []*models.AccountInfo `json:"items"`
	Page       uint                  `json:"page"`
	PageSize   uint                  `json:"page_size"`
	Total      uint                  `json:"total"`
	TotalPages uint                  `json:"total_pages"`
}

type accountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) AccountRepository {
	return &accountRepository{
		db: db,
	}
}

func (a accountRepository) GetListAccounts(ctx context.Context, q string, page uint, limit uint) (AccountPage, error) {
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	var mods []qm.QueryMod
	if q != "" {
		mods = append(mods, qm.Where("cAccName LIKE ?", q+"%"))
	}
	total64, err := models.AccountInfos(mods...).Count(ctx, a.db)
	if err != nil {
		return AccountPage{}, err
	}
	offset := (page - 1) * limit
	mods = append(mods, qm.OrderBy("cAccName"), qm.Limit(int(limit)), qm.Offset(int(offset)))
	accounts, err := models.AccountInfos(mods...).All(ctx, a.db)
	if err != nil {
		return AccountPage{}, err
	}
	total := uint(total64)
	pages := total / limit
	if total%limit != 0 {
		pages++
	}
	return AccountPage{Items: accounts, Page: page, PageSize: limit, Total: total, TotalPages: pages}, nil
}

func (a accountRepository) GetAccount(ctx context.Context, account string) (*models.AccountInfo, error) {
	return models.AccountInfos(qm.Where("cAccName = ?", account)).One(ctx, a.db)
}

func (a accountRepository) CreateAccount(ctx context.Context, account, password string) error {
	_, err := a.db.ExecContext(ctx, "insert into Account_Info(cAccName,cPassword,iClientID,bIsBanned) values(@p1,@p2,0,0)", account, password)
	return err
}

func (a accountRepository) VerifyPlayer(ctx context.Context, account, password string) (bool, error) {
	var stored string
	var banned bool
	err := a.db.QueryRowContext(ctx, "select cPassword,isnull(bIsBanned,0) from Account_Info where cAccName=@p1", account).Scan(&stored, &banned)
	if err != nil {
		return false, err
	}
	return !banned && stored == password, nil
}

func (a accountRepository) ChangePassword(ctx context.Context, account, current, next string) error {
	result, err := a.db.ExecContext(ctx, "update Account_Info set cPassword=@p1 where cAccName=@p2 and cPassword=@p3 and isnull(bIsBanned,0)=0", next, account, current)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
func (a accountRepository) ResetPassword(ctx context.Context, account, password string) error {
	_, err := a.db.ExecContext(ctx, "update Account_info set cPassword=@p1 where cAccName=@p2", password, account)
	return err
}

func (a accountRepository) Block(ctx context.Context, account string) error {
	_, err := a.db.ExecContext(ctx, "update Account_Info set iClientID=-1, bIsBanned=1 where cAccName=@p1", account)
	return err
}

func (a accountRepository) Unblock(ctx context.Context, account string) error {
	_, err := a.db.ExecContext(ctx, "update Account_Info set iClientID=0, bIsBanned=0 where cAccName=@p1", account)
	return err
}

func (a accountRepository) Kick(ctx context.Context, account string) error {
	_, err := a.db.ExecContext(ctx, "update Account_info set iClientID=0 where cAccName=@p1", account)
	return err
}
