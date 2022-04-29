package interfaces

import (
	"context"
	"net/http"

	"github.com/tim3-p/go-ya-diplom/internal/models"
)

type User interface {
	Create(ctx context.Context, user models.User) error
	GetByLogin(ctx context.Context, login string) (models.User, error)
}

type Order interface {
	Create(ctx context.Context, order models.Order) error
	GetByUserID(ctx context.Context, userID uint64) ([]models.Order, error)
	GetByNumber(ctx context.Context, number string) (models.Order, error)
}

type Withdrawal interface {
	Create(ctx context.Context, withdrawal models.Withdrawal) error
	GetByUserID(ctx context.Context, userID uint64) ([]models.Withdrawal, error)
}

type CookieAuthenticator interface {
	SetCookie(w http.ResponseWriter, login string) error
}

type PointAccrualService interface {
	Accrue(order string)
}

type Middleware interface {
	Handle(next http.HandlerFunc) http.HandlerFunc
}
