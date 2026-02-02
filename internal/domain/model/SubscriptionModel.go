package model

type Subscription struct {
	ID       int
	UserID   int64
	LotName  string
	MinPrice float64
	Category string
}
