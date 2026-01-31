package models

type Lot struct {
	ID       string
	Title    string
	Price    float64
	Category string
}

type Subscription struct {
	ID       int64
	ChatID   int64
	Category int
	Query    string
	MaxPrice float64
}
