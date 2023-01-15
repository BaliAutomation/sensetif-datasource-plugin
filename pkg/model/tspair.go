package model

import "time"

type TsPair struct {
    TS    time.Time `json:"ts"`
    Value float64   `json:"value"`
}
