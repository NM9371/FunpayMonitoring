package port

import "context"

type Notifier interface {
	Notify(ctx context.Context, userID int64, message string) error
}
