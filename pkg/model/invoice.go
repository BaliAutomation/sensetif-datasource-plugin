package model

import "time"

type Invoice struct {
	InvoiceDate     time.Time `json:"invoicedate"`
	PlanTitle       string    `json:"plantitle"`
	PlanDescription string    `json:"plandescription"`
	Stats           string    `json:"stats"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
}
