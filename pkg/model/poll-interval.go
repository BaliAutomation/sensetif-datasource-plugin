package model

type PollInterval string

// PollInterval values
const (
	One_minute      PollInterval = "one_minute"
	Five_minutes    PollInterval = "five_minutes"
	Ten_minutes     PollInterval = "ten_minutes"
	Fifteen_minutes PollInterval = "fifteen_minutes"
	Twenty_minutes  PollInterval = "twenty_minutes"
	Thirty_minutes  PollInterval = "thirty_minutes"
	One_hour        PollInterval = "one_hour"
	Two_hours       PollInterval = "two_hours"
	Three_hours     PollInterval = "three_hours"
	Six_hours       PollInterval = "six_hours"
	Twelve_hours    PollInterval = "twelve_hours"
	One_day         PollInterval = "one_day"
	Weekly          PollInterval = "weekly"
	Monthly         PollInterval = "monthly"
)
