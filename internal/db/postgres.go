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
	Name  string
	Price float64
	URL   string
}

type Subscription struct {
	UserID   int64
	LotName  string
	MinPrice float64
	URL      string
}

func NewPostgres() (*Postgres, error) {
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "postgres"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", host, user, password, dbName)
	var db *sql.DB
	var err error
	for i := 0; i < 10; i++ { // 10 попыток
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

// Получение всех подписок
func (p *Postgres) GetSubscriptions() ([]Subscription, error) {
	rows, err := p.db.Query("SELECT user_id, lot_name, min_price, url FROM subscriptions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []Subscription
	for rows.Next() {
		var s Subscription
		if err := rows.Scan(&s.UserID, &s.LotName, &s.MinPrice, &s.URL); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	return subs, nil
}

// Вставка новой подписки
func (p *Postgres) InsertSubscription(sub Subscription) error {
	query := `
	INSERT INTO subscriptions (user_id, lot_name, min_price, url)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (user_id, lot_name, url) DO NOTHING
	`
	_, err := p.db.Exec(query, sub.UserID, sub.LotName, sub.MinPrice, sub.URL)
	return err
}

// Вставка истории цены
func (p *Postgres) InsertPriceHistory(lot Lot) error {
	query := `
	INSERT INTO price_history (name, price, url)
	VALUES ($1, $2, $3)
	`
	_, err := p.db.Exec(query, lot.Name, lot.Price, lot.URL)
	return err
}
