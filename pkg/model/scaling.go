package model

type Scaling string

// Scaling values
//goland:noinspection GoUnusedConst
const (
	// Lin out = k * x + m
	Lin Scaling = "lin"

	// Ln out = k * ln(m*x)
	Ln Scaling = "ln"

	// Exp out = k * e^(m*x)
	Exp Scaling = "exp"

	// Rad Inputs are degrees, to be converted to radians.
	Rad Scaling = "rad"

	// Deg Input are radians, to be converted to degrees.
	Deg Scaling = "deg"

	// FtoC Input Fahrenheit, output Celsius
	FtoC Scaling = "fToC"

	// CtoF Input Celsius, output Fahrenheit
	CtoF Scaling = "cToF"

	// KtoC Input Kelvin, output Celsius
	KtoC Scaling = "kToC"

	// CtoK Input Celsius, output Kelvin
	CtoK Scaling = "cToK"

	// FtoK  Input Fahrenheit, output Kelvin
	FtoK Scaling = "fToK"

	// KtoF Input Kelvin, output Fahrenheit
	KtoF Scaling = "kToF"
)
