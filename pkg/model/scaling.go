package model

type Scaling int

// Scaling values
//goland:noinspection GoUnusedConst
const (
	// Lin out = k * x + m
	Lin Scaling = 0

	// Ln out = k * ln(m*x)
	Ln Scaling = 1

	// Exp out = k * e^(m*x)
	Exp Scaling = 2

	// Rad Inputs are degrees, to be converted to radians.
	Rad Scaling = 3

	// Deg Input are radians, to be converted to degrees.
	Deg Scaling = 4

	// FtoC Input Fahrenheit, output Celsius
	FtoC Scaling = 5

	// CtoF Input Celsius, output Fahrenheit
	CtoF Scaling = 6

	// KtoC Input Kelvin, output Celsius
	KtoC Scaling = 7

	// CtoK Input Celsius, output Kelvin
	CtoK Scaling = 8

	// FtoK  Input Fahrenheit, output Kelvin
	FtoK Scaling = 9

	// KtoF Input Kelvin, output Fahrenheit
	KtoF Scaling = 10
)
