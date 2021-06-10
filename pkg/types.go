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

// // Incoming Form Types
// type DatapointSettings struct {
// 	Name     string // validate regexp [a-z][A-Za-z0-9_.]*
// 	Interval PollInterval
// 	URL      string // validate URL, incl anchor and query arguments, but disallow user pwd@

// 	//   // Authentication is going to need a lot in the future, but for now user/pass is fine
// 	AuthenticationType AuthenticationType
// 	Credentials        Object // if Authentication == userpass, then this is "{ user  "username", password  "pass" }"

// 	Format          OriginDocumentFormat
// 	ValueExpression string // if format==xml, then xpath. if format==json, then jsonpath. If there is library available for validation, do that. If not, put in a function and we figure that out later.
// 	Unit            string // Allow all characters

// 	// Ideally only show k and m for ScalingFunctions that uses them, and show the formula with the scaling function
// 	Scaling ScalingFunction
// 	K       number
// 	M       number

// 	TimestampType       TimestampType
// 	TimestampExpression string // if format==xml, then xpath. if format==json, then jsonpath.
// 	TimeToLive          TimeToLive
// }
// type SubsystemSettings struct {
// 	Name          string // validate regexp [a-z][A-Za-z0-9_]*
// 	Title         string // allow all characters
// 	Locallocation string // allow all characters
// 	Datapoints    []DatapointSettings
// }

type ProjectSettings struct {
	Name        string // validate regexp[a-z][A-Za-z0-9_]*
	Title       string // allow all characters
	City        string // allow all characters
	Country     string // country list?
	Timezone    string // "UTC" or {continent}/{city}, ex Europe/Stockholm
	Geolocation string // geo coordinates
	// Subsystems  []SubsystemSettings
}
