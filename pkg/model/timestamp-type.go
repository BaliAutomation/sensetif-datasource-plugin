package model

type TimestampType int

// TimestampType values
const (
	EpochMillis    TimestampType = 0
	EpochSeconds   TimestampType = 1
	ISO8601_zoned  TimestampType = 2
	ISO8601_offset TimestampType = 3
)
