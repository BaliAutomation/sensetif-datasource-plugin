package model

import (
	"time"
)

type Payment struct {
	InvoiceDate time.Time `json:"invoicedate"`
	PaymentDate time.Time `json:"paymentdate"`
	Amount      float32   `json:"amount"`
	Currency    string    `json:"currency"`
}
