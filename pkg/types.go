package main

import (
	"time"
)

type SensorRef struct {
	project   string
	subsystem string
	datapoint string
}

type TsPair struct {
	ts    time.Time
	value float64
}
