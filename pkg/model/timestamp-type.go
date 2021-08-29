package model

type TimestampType string

// TimestampType values
const (
	EpochMillis    TimestampType = "epochMillis"
	EpochSeconds   TimestampType = "epochSeconds"
	ISO8601_zoned  TimestampType = "iso8601_zoned"
	ISO8601_offset TimestampType = "iso8601_offset"
	PollTime       TimestampType = "polltime"
)
