package model

type TimeToLive string

// TimeToLive values
const (
	A TimeToLive = "a" // 10 days
	B TimeToLive = "b" // 40 days
	C TimeToLive = "c" // 100 days
	D TimeToLive = "d" // 200 days
	E TimeToLive = "e" // 400 days
	F TimeToLive = "f" // 750 days
	G TimeToLive = "g" // 1200 days
	H TimeToLive = "h" // 1500 days
	I TimeToLive = "i" // 1900 days
	J TimeToLive = "j" // 3700 days
	K TimeToLive = "k" // forever
)

var (
	TimeToLives = []TimeToLive{
		A,
		B,
		C,
		D,
		E,
		F,
		G,
		H,
		I,
		J,
		K,
	}
)
