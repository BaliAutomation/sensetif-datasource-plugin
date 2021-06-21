package model

type PollInterval int

// PollInterval values
const (
	One_minute      PollInterval = 0
	Five_minutes    PollInterval = 1
	Ten_minutes     PollInterval = 2
	Fifteen_minutes PollInterval = 3
	Twenty_minutes  PollInterval = 4
	Thirty_minutes  PollInterval = 5
	One_hour        PollInterval = 6
	Two_hours       PollInterval = 7
	Three_hours     PollInterval = 8
	Six_hours       PollInterval = 9
	Twelve_hours    PollInterval = 10
	One_day         PollInterval = 11
	Weekly          PollInterval = 12
	Monthly         PollInterval = 13
)
