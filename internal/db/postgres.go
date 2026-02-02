package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/NM9371/FunpayMonitoring/internal/domain/model"
	_ "github.com/lib/pq"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgres() (*Postgres, error) {
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "postgres"
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s sslmode=disable",
		host, user, password, dbName,
	)

	var db *sql.DB
	var err error

	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", dsn)
		if err == nil && db.Ping() == nil {
			break
		}
		log.Println("Waiting for Postgres to be ready...")
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return nil, err
	}

	log.Println("✅ Connected to PostgreSQL")
	return &Postgres{db: db}, nil
}

func (p *Postgres) ListAll(ctx context.Context) ([]model.Subscription, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, user_id, lot_name, min_price, category
		FROM subscriptions
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []model.Subscription
	for rows.Next() {
		var s model.Subscription
		if err := rows.Scan(&s.ID, &s.UserID, &s.LotName, &s.MinPrice, &s.Category); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	return subs, nil
}

func (p *Postgres) ListByUser(ctx context.Context, userID int64) ([]model.Subscription, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, user_id, lot_name, min_price, category
		FROM subscriptions
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []model.Subscription
	for rows.Next() {
		var s model.Subscription
		if err := rows.Scan(&s.ID, &s.UserID, &s.LotName, &s.MinPrice, &s.Category); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	return subs, nil
}

func (p *Postgres) Add(ctx context.Context, sub model.Subscription) error {
	query := `
		INSERT INTO subscriptions (user_id, lot_name, min_price, category)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, lot_name, category) DO NOTHING
	`
	_, err := p.db.ExecContext(ctx, query, sub.UserID, sub.LotName, sub.MinPrice, sub.Category)
	return err
}

func (p *Postgres) Remove(ctx context.Context, userID int64, category string, lotName string) error {
	_, err := p.db.ExecContext(
		ctx,
		`DELETE FROM subscriptions WHERE user_id = $1 AND category = $2 AND lot_name = $3`,
		userID, category, lotName,
	)
	return err
}
