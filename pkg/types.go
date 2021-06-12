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

// Incoming Form Types
type DatapointSettings struct {
	Project   string       `json:"project"`
	Subsystem string       `json:"subsystem"`
	Name      string       `json:"name"` // validate regexp [a-z][A-Za-z0-9_.]*
	Interval  PollInterval `json:"interval"`
	URL       string       `json:"url"` // validate URL, incl anchor and query arguments, but disallow user pwd@

	// Authentication is going to need a lot in the future, but for now user/pass is fine
	AuthenticationType AuthenticationType `json:"authenticationType"`
	Credentials        Auth               `json:"credentials"` // if Authentication == basic, then this is BasicAuth

	Format          OriginDocumentFormat `json:"format"`
	ValueExpression string               `json:"valueExpression"` // if format==xml, then xpath. if format==json, then jsonpath. If there is library available for validation, do that. If not, put in a function and we figure that out later.
	Unit            string               `json:"unit"`            // Allow all characters

	// Ideally only show k and m for ScalingFunctions that uses them, and show the formula with the scaling function
	Scaling ScalingFunction `json:"scaling"`
	K       float64         `json:"k"`
	M       float64         `json:"m"`

	TimestampType       TimestampType `json:"timestampType"`
	TimestampExpression string        `json:"timestampExpression"` // if format==xml, then xpath. if format==json, then jsonpath.
	TimeToLive          TimeToLive    `json:"timeToLive"`
}

type SubsystemSettings struct {
	Project       string `json:"project"`
	Name          string `json:"name"`          // validate regexp [a-z][A-Za-z0-9_]*
	Title         string `json:"title"`         // allow all characters
	Locallocation string `json:"locallocation"` // allow all characters
}

type ProjectSettings struct {
	Name        string `json:"name"`        // validate regexp[a-z][A-Za-z0-9_]*
	Title       string `json:"title"`       // allow all characters
	City        string `json:"city"`        // allow all characters
	Country     string `json:"country"`     // country list?
	Timezone    string `json:"timezone"`    // "UTC" or {continent}/{city}, ex Europe/Stockholm
	Geolocation string `json:"geolocation"` // geo coordinates
}

type PollInterval int

type AuthenticationType int

type Auth struct{}

type BasicAuth struct {
	Auth
	User     string `json:"user"`
	Password string `json:"password"`
}

type QueryParamAuth struct {
	Auth
	Key   string `json:"key"`
	Value string `json:"value"`
}

type OriginDocumentFormat int

type TimestampType int

type TimeToLive int

type ScalingFunction int

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

// AuthenticationType values
const (
	none       AuthenticationType = 0
	basic      AuthenticationType = 1
	queryParam AuthenticationType = 2
)

// OriginDocumentFormat values
const (
	json OriginDocumentFormat = 0
	xml  OriginDocumentFormat = 1
)

// TimestampType values
const (
	epochMillis    TimestampType = 0
	epochSeconds   TimestampType = 1
	iso8601_zoned  TimestampType = 2
	iso8601_offset TimestampType = 3
)

// TimeToLive values
const (
	a TimeToLive = 0 // 100 days
	b TimeToLive = 1 // 200 days
	c TimeToLive = 2 // 400 days
	d TimeToLive = 3 // 750 days
	e TimeToLive = 4 // 1100 days
	f TimeToLive = 5 // 1500 days
	g TimeToLive = 6 // 1900 days
	h TimeToLive = 7 // forever
)

// ScalingFunction values
const (
	/**
	 * out = k * x + m
	 */
	lin ScalingFunction = 0

	/**
	 * out = k * ln(m*x)
	 */
	ln ScalingFunction = 1

	/**
	 * out = k * e^(m*x)
	 */
	exp ScalingFunction = 2

	/**
	 * Inputs are degrees, to be converted to radians.
	 */
	rad ScalingFunction = 3

	/**
	 * Input are radians, to be converted to degrees.
	 */
	deg ScalingFunction = 4

	/**
	 * Input Fahrenheit, output Celsius
	 */
	fToC ScalingFunction = 5

	/**
	 * Input Celsius, output Fahrenheit
	 */
	cToF ScalingFunction = 6

	/**
	 * Input Kelvin, output Celsius
	 */
	kToC ScalingFunction = 7

	/**
	 * Input Celsius, output Kelvin
	 */
	cToK ScalingFunction = 8

	/**
	 * Input Fahrenheit, output Kelvin
	 */
	fToK ScalingFunction = 9

	/**
	 * Input Kelvin, output Fahrenheit
	 */
	kTof ScalingFunction = 10
)
