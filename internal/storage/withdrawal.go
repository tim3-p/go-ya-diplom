package storage

import (
	"context"
	"database/sql"
	"errors"
	"sort"

	"github.com/tim3-p/go-ya-diplom/internal/models"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
)

type Withdrawal struct {
	db *sql.DB
}

func CreateWithdrawal(db *sql.DB) *Withdrawal {
	return &Withdrawal{
		db: db,
	}
}

func (r *Withdrawal) Create(ctx context.Context, withdrawal models.Withdrawal) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
	}()

	var balance float64
	row := tx.QueryRowContext(ctx, `SELECT balance FROM users WHERE id = $1`, withdrawal.UserID)
	err = row.Scan(&balance)
	if err != nil {
		return err
	}

	if balance < withdrawal.Sum {
		return ErrInsufficientBalance
	}

	createWithdrawalStatement := `INSERT INTO withdrawal ("order", sum, created_at, user_id) VALUES ($1, $2, $3, $4)`
	_, err = tx.ExecContext(ctx, createWithdrawalStatement, withdrawal.Order, withdrawal.Sum, withdrawal.CreatedAt, withdrawal.UserID)
	if err != nil {
		return err
	}

	updateBalanceStatement := `UPDATE users SET balance = balance - $1, withdrawn = withdrawn + $1 WHERE id = $2`
	_, err = tx.ExecContext(ctx, updateBalanceStatement, withdrawal.Sum, withdrawal.UserID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Withdrawal) GetByUserID(ctx context.Context, userID uint64) ([]models.Withdrawal, error) {
	var withdrawals []models.Withdrawal

	rows, err := r.db.QueryContext(ctx, `SELECT id, "order", sum, created_at, user_id FROM withdrawal WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var withdrawal models.Withdrawal
		err := rows.Scan(&withdrawal.ID, &withdrawal.Order, &withdrawal.Sum, &withdrawal.CreatedAt, &withdrawal.UserID)
		if err != nil {
			return nil, err
		}

		withdrawals = append(withdrawals, withdrawal)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	if len(withdrawals) == 0 {
		return nil, sql.ErrNoRows
	}

	sort.Slice(withdrawals, func(i, j int) bool {
		return withdrawals[i].CreatedAt.Before(withdrawals[j].CreatedAt)
	})

	return withdrawals, nil
}
