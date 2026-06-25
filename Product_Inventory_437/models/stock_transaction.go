package models

import "time"

type StockTransaction struct {
	ID        int64     `json:"id"`
	ProductID int64     `json:"product_id"`
	Type      string    `json:"type"`
	Quantity  int       `json:"quantity"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}