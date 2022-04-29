package storage

import (
	"context"
	"database/sql"
	"sort"

	"github.com/tim3-p/go-ya-diplom/internal/models"
)

type Order struct {
	db *sql.DB
}

func CreateOrder(db *sql.DB) *Order {
	return &Order{
		db: db,
	}
}

func (r *Order) Create(ctx context.Context, order models.Order) error {
	sqlStatement := `INSERT INTO "order" (number, status, uploaded_at, user_id) VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, sqlStatement, order.Number, order.Status, order.CreatedAt, order.UserID)
	return err
}

func (r *Order) GetByUserID(ctx context.Context, userID uint64) ([]models.Order, error) {
	var orders []models.Order

	rows, err := r.db.QueryContext(ctx, `SELECT id, number, status, accrual, uploaded_at, user_id FROM "order" WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var order models.Order
		err := rows.Scan(&order.ID, &order.Number, &order.Status, &order.Accrual, &order.CreatedAt, &order.UserID)
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	if len(orders) == 0 {
		return nil, sql.ErrNoRows
	}

	sort.Slice(orders, func(i, j int) bool {
		return orders[i].CreatedAt.Before(orders[j].CreatedAt)
	})

	return orders, nil
}

func (r *Order) GetByNumber(ctx context.Context, number string) (models.Order, error) {
	var order models.Order

	sqlStatement := `SELECT id, number, status, accrual, uploaded_at, user_id FROM "order" WHERE number = $1`
	row := r.db.QueryRowContext(ctx, sqlStatement, number)
	err := row.Scan(&order.ID, &order.Number, &order.Status, &order.Accrual, &order.CreatedAt, &order.UserID)
	if err != nil {
		return models.Order{}, err
	}

	return order, nil
}

func (r *Order) UpdateAccrual(ctx context.Context, accrual models.Accrual) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	updateOrderStatement := `UPDATE "order" SET status = $1, accrual = $2 WHERE number = $3`
	_, err = tx.ExecContext(ctx, updateOrderStatement, accrual.Status, accrual.Accrual, accrual.Order)
	if err != nil {
		return err
	}

	increaseBalanceStatement := `
UPDATE "user"
SET balance = "user".balance + $1
FROM "user" as u
INNER JOIN "order" ON u.id = "order".user_id
WHERE "order".number = $2
`
	_, err = tx.ExecContext(ctx, increaseBalanceStatement, accrual.Accrual, accrual.Order)
	if err != nil {
		return err
	}

	return tx.Commit()
}
