package port

import (
	"context"

	"github.com/NM9371/FunpayMonitoring/internal/domain/model"
)

type LotProvider interface {
	FindLots(ctx context.Context, category string, query string) ([]model.Lot, error)
}
