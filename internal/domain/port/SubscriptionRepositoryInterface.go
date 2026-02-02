package port

import (
	"context"

	"github.com/NM9371/FunpayMonitoring/internal/domain/model"
)

type SubscriptionRepository interface {
	ListAll(ctx context.Context) ([]model.Subscription, error)
	ListByUser(ctx context.Context, userID int64) ([]model.Subscription, error)
	Add(ctx context.Context, sub model.Subscription) error
	Remove(ctx context.Context, userID int64, category string, lotName string) error
}
