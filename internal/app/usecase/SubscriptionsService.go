package usecase

import (
	"context"

	"github.com/NM9371/FunpayMonitoring/internal/domain/model"
	"github.com/NM9371/FunpayMonitoring/internal/domain/port"
)

type SubscriptionsService struct {
	repo port.SubscriptionRepository
}

func NewSubscriptionsService(repo port.SubscriptionRepository) *SubscriptionsService {
	return &SubscriptionsService{repo: repo}
}

func (s *SubscriptionsService) Add(ctx context.Context, sub model.Subscription) error {
	return s.repo.Add(ctx, sub)
}

func (s *SubscriptionsService) ListByUser(ctx context.Context, userID int64) ([]model.Subscription, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *SubscriptionsService) Remove(ctx context.Context, userID int64, category string, lotName string) error {
	return s.repo.Remove(ctx, userID, category, lotName)
}
