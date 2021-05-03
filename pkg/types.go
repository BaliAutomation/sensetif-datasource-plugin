package main

import (
	"time"
)

type SensorRef struct {
	project   string
	subsystem string
	sensor    string
}

type TsPair struct {
	ts    time.Time
	value float64
}
