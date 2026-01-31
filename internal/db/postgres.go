package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type Postgres struct {
	db *sql.DB
}

type Lot struct {
	Name     string
	Price    float64
	Category string
}

type Subscription struct {
	ID       int
	UserID   int64
	LotName  string
	MinPrice float64
	Category string
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

func (p *Postgres) GetSubscriptions() ([]Subscription, error) {
	rows, err := p.db.Query(`
		SELECT id, user_id, lot_name, min_price, category
		FROM subscriptions
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []Subscription
	for rows.Next() {
		var s Subscription
		if err := rows.Scan(&s.ID, &s.UserID, &s.LotName, &s.MinPrice, &s.Category); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	return subs, nil
}

func (p *Postgres) InsertSubscription(sub Subscription) error {
	query := `
		INSERT INTO subscriptions (user_id, lot_name, min_price, category)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, lot_name, category) DO NOTHING
	`
	_, err := p.db.Exec(query, sub.UserID, sub.LotName, sub.MinPrice, sub.Category)
	return err
}

func (p *Postgres) DeleteSubscription(userID int64, lotName string) error {
	_, err := p.db.Exec(
		`DELETE FROM subscriptions WHERE user_id = $1 AND lot_name = $2`,
		userID, lotName,
	)
	return err
}

func (p *Postgres) InsertPriceHistory(lot Lot) error {
	_, err := p.db.Exec(
		`INSERT INTO price_history (category, lot_name, price) VALUES ($1, $2, $3)`,
		lot.Category, lot.Name, lot.Price,
	)
	return err
}
