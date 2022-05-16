package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/tim3-p/go-ya-diplom/internal/models"
)

type Order interface {
	GetByNumber(ctx context.Context, number string) (models.Order, error)
	UpdateAccrual(ctx context.Context, accrual models.Accrual) error
}

type Accrual struct {
	orders               chan string
	accrualSystemAddress string
	order                Order
}

func NewAccrual(
	accrualSystemAddress string,
	order Order,
) *Accrual {
	return &Accrual{
		orders:               make(chan string, 100),
		accrualSystemAddress: accrualSystemAddress,
		order:                order,
	}
}

func (s *Accrual) Start() {
	go func() {
		for order := range s.orders {
			err := s.handleOrder(order)
			if err != nil {
				s.Accrue(order)
			}
		}
	}()
}

func (s *Accrual) handleOrder(order string) error {
	url := fmt.Sprintf("%s/api/orders/%s", s.accrualSystemAddress, order)
	response, err := http.Get(url)
	if err != nil {
		return err
	}

	switch response.StatusCode {
	case http.StatusOK:
		defer response.Body.Close()
		payload, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}

		accrual := models.Accrual{}
		if err := json.Unmarshal(payload, &accrual); err != nil {
			return err
		}

		err = s.order.UpdateAccrual(context.Background(), accrual)
		if err != nil {
			return err
		}
	case http.StatusTooManyRequests:
		s.Accrue(order)
	case http.StatusInternalServerError:
		s.Accrue(order)
	}

	return nil
}

func (s *Accrual) Stop() {
	close(s.orders)
}

func (s *Accrual) Accrue(orderNumber string) {
	s.orders <- orderNumber
}
