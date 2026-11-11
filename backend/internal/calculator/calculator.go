// Package calculator evaluates arithmetic expressions.
package calculator

import (
	"errors"
	"math"
)

// Errors for expressions that are valid but have no usable answer. Match them with errors.Is.
var (
	ErrDivisionByZero       = errors.New("division by zero")
	ErrNegativeSquareRoot   = errors.New("square root of a negative number")
	ErrNonPositiveLogarithm = errors.New("logarithm of a number that is not positive")
	ErrOutOfRange           = errors.New("result is out of range")
	ErrUndefinedResult      = errors.New("result is not a real number")
)

type (
	unaryFunc  func(a float64) (float64, error)
	binaryFunc func(a, b float64) (float64, error)
)

func add(a, b float64) (float64, error)      { return a + b, nil }
func subtract(a, b float64) (float64, error) { return a - b, nil }
func multiply(a, b float64) (float64, error) { return a * b, nil }
func negate(a float64) (float64, error)      { return -a, nil }
func percent(a float64) (float64, error)     { return a / 100, nil }

func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, ErrDivisionByZero
	}
	return a / b, nil
}

func power(a, b float64) (float64, error) {
	if a == 0 && b < 0 {
		return 0, ErrDivisionByZero
	}
	return math.Pow(a, b), nil
}

func squareRoot(a float64) (float64, error) {
	if a < 0 {
		return 0, ErrNegativeSquareRoot
	}
	return math.Sqrt(a), nil
}

// sine and cosine take degrees. Multiples of 90° return exact values, so sin(180) is 0
// rather than 1.2e-16.
func sine(a float64) (float64, error) {
	if math.Mod(a, 180) == 0 {
		return 0, nil
	}
	return math.Sin(radians(a)), nil
}

func cosine(a float64) (float64, error) {
	if math.Abs(math.Mod(a, 180)) == 90 {
		return 0, nil
	}
	return math.Cos(radians(a)), nil
}

// radians reduces a to one turn first, so large angles keep their precision.
func radians(degrees float64) float64 {
	return math.Mod(degrees, 360) * math.Pi / 180
}

func logarithm(a float64) (float64, error) {
	if a <= 0 {
		return 0, ErrNonPositiveLogarithm
	}
	return math.Log10(a), nil
}

func exponential(a float64) (float64, error) { return math.Exp(a), nil }

// finite turns NaN and ±Inf into errors and -0 into 0.
func finite(result float64, err error) (float64, error) {
	switch {
	case err != nil:
		return 0, err
	case math.IsNaN(result):
		return 0, ErrUndefinedResult
	case math.IsInf(result, 0):
		return 0, ErrOutOfRange
	case result == 0:
		return 0, nil
	}
	return result, nil
}
