package model

type ScalingFunction int

// ScalingFunction values
const (
	/**
	 * out = k * x + m
	 */
	Lin ScalingFunction = 0

	/**
	 * out = k * ln(m*x)
	 */
	Ln ScalingFunction = 1

	/**
	 * out = k * e^(m*x)
	 */
	Exp ScalingFunction = 2

	/**
	 * Inputs are degrees, to be converted to radians.
	 */
	Rad ScalingFunction = 3

	/**
	 * Input are radians, to be converted to degrees.
	 */
	Deg ScalingFunction = 4

	/**
	 * Input Fahrenheit, output Celsius
	 */
	FtoC ScalingFunction = 5

	/**
	 * Input Celsius, output Fahrenheit
	 */
	CtoF ScalingFunction = 6

	/**
	 * Input Kelvin, output Celsius
	 */
	KtoC ScalingFunction = 7

	/**
	 * Input Celsius, output Kelvin
	 */
	CtoK ScalingFunction = 8

	/**
	 * Input Fahrenheit, output Kelvin
	 */
	FtoK ScalingFunction = 9

	/**
	 * Input Kelvin, output Fahrenheit
	 */
	Ktof ScalingFunction = 10
)
