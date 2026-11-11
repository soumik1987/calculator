package calculator_test

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/soumik1987/sami-calculator/backend/internal/calculator"
)

func TestEvaluate(t *testing.T) {
	tests := []struct {
		expr string
		want float64
	}{
		{"(986x(4+8))+2", 11834},
		{"(986×(4+8))+√16", 11836},
		{"(986*(4+8))+sqrt(16)", 11836},
		{"42", 42},
		{"  1 +   2  ", 3},

		{"18+4", 22},
		{"-5+-7", -12},
		{"0.5+0.25", 0.75},
		{"0.1+0.2", 0.30000000000000004},
		{"18-4", 14},
		{"4-18", -14},
		{"4--6", 10},
		{"3 − 1", 2},
		{"-3*4", -12},
		{"-3*-4", 12},
		{"1.5*0.5", 0.75},
		{"9/4", 2.25},
		{"9/-3", -3},
		{"0/5", 0},
		{"1/0.25", 4},
		{"1/3", 1.0 / 3},
		{"2^8", 256},
		{"5^0", 1},
		{"2^-1", 0.5},
		{"9^0.5", 3},
		{"(-2)^2", 4},
		{"(-2)^3", -8},
		{"0^3", 0},
		{"√81", 9},
		{"√0", 0},
		{"√0.25", 0.5},
		{"√2", math.Sqrt2},
		{"25%", 0.25},
		{"-50%", -0.5},
		{"250%", 2.5},

		{"2+3*4", 14},
		{"(2+3)*4", 20},
		{"10-2-3", 5},
		{"8÷2×4", 16},
		{"2^3^2", 512},
		{"-2^2", -4},
		{"--3", 3},
		{"+3", 3},
		{"200*10%", 20},
		{"√16+√9", 7},
		{"√(9+16)", 5},
		{"√√16", 2},
		{"√16^2", 16},
		{"sin(0)", 0},
		{"sin(90)", 1},
		{"sin(180)", 0},
		{"sin(-360)", 0},
		{"sin(450)", 1},
		{"cos(0)", 1},
		{"cos(90)", 0},
		{"cos(-90)", 0},
		{"cos(180)", -1},
		{"cos(270)", 0},
		{"log(1000)", 3},
		{"log(1)", 0},
		{"log(0.01)", -2},
		{"exp(0)", 1},
		{"exp(1)", math.E},
		{"exp(-1000)", 0},
		{"sin 90", 1},
		{"SIN(90)", 1},
		{"sin(90)^2", 1},
		{"2×cos(0)", 2},
		{"-cos(0)", -1},
		{"log(exp(0))", 0},
		{"√exp(0)", 1},
		{"1e3+1", 1001},
		{"1e+2", 100},
		{"5E-1", 0.5},
		{".5+.5", 1},

		{"1e-200*1e-200", 0},
		{"-1*0", 0},
		{"0/-5", 0},
		{"-0", 0},
		{"(-0)^3", 0},

		{strings.Repeat("(", 249) + "1" + strings.Repeat(")", 249), 1},
		{strings.Repeat("-", 498) + "1", 1},
		{strings.Repeat("√", 499) + "1", 1},
		{strings.Repeat("1^", 249) + "1", 1},
	}

	for _, tt := range tests {
		t.Run(shortName(tt.expr), func(t *testing.T) {
			got, err := calculator.Evaluate(tt.expr)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want || math.Signbit(got) != math.Signbit(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluateSyntaxErrors(t *testing.T) {
	tests := []struct {
		expr string
		want string
	}{
		{"", "Enter an expression."},
		{"   ", "Enter an expression."},
		{"2+", "The expression is incomplete."},
		{"2^", "The expression is incomplete."},
		{"-", "The expression is incomplete."},
		{"(2+", "The expression is incomplete."},
		{"√", "The expression is incomplete."},
		{"2e", `Unknown word "e" at position 2.`},
		{"(2+3", `Missing ")" for the "(" at position 1.`},
		{"2+3)", `Unexpected ")" at position 4.`},
		{"2(3)", `Unexpected "(" at position 2.`},
		{"2 3", `Unexpected "3" at position 3.`},
		{"*2", `Unexpected "*" at position 1.`},
		{"2+abc", `Unknown word "abc" at position 3.`},
		{"tan(1)", `Unknown word "tan" at position 1.`},
		{"sin", "The expression is incomplete."},
		{"sin+1", `Unexpected "+" at position 4.`},
		{"2sin(1)", `Unexpected "sin" at position 2.`},
		{"2$3", `Unexpected character "$" at position 2.`},
		{"1.2.3", `Invalid number "1.2.3" at position 1.`},
		{"1e400", `Invalid number "1e400" at position 1.`},
		{strings.Repeat("1", 501), "The expression is longer than 500 characters."},
	}

	for _, tt := range tests {
		t.Run(shortName(tt.expr), func(t *testing.T) {
			_, err := calculator.Evaluate(tt.expr)

			var syntaxErr *calculator.SyntaxError
			if !errors.As(err, &syntaxErr) {
				t.Fatalf("got %v, want *SyntaxError", err)
			}
			if err.Error() != tt.want {
				t.Errorf("got message %q, want %q", err.Error(), tt.want)
			}
		})
	}
}

func TestEvaluateMathErrors(t *testing.T) {
	tests := []struct {
		expr string
		want error
	}{
		{"1/0", calculator.ErrDivisionByZero},
		{"0/0", calculator.ErrDivisionByZero},
		{"-1/0", calculator.ErrDivisionByZero},
		{"1/0+1", calculator.ErrDivisionByZero},
		{"1+1/0", calculator.ErrDivisionByZero},
		{"-(1/0)", calculator.ErrDivisionByZero},
		{"1/(2-2)", calculator.ErrDivisionByZero},
		{"0^-1", calculator.ErrDivisionByZero},
		{"√(-4)", calculator.ErrNegativeSquareRoot},
		{"√(-0.01)", calculator.ErrNegativeSquareRoot},
		{"log(0)", calculator.ErrNonPositiveLogarithm},
		{"log(-1)", calculator.ErrNonPositiveLogarithm},
		{"log(-0)", calculator.ErrNonPositiveLogarithm},
		{"exp(1000)", calculator.ErrOutOfRange},
		{"(-8)^0.5", calculator.ErrUndefinedResult},
		{"10^1000", calculator.ErrOutOfRange},
		{"1e308+1e308", calculator.ErrOutOfRange},
		{"-1e308-1e308", calculator.ErrOutOfRange},
		{"1e308*10", calculator.ErrOutOfRange},
		{"1e308/0.1", calculator.ErrOutOfRange},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			_, err := calculator.Evaluate(tt.expr)
			if !errors.Is(err, tt.want) {
				t.Errorf("got error %v, want %v", err, tt.want)
			}
		})
	}
}

func shortName(expr string) string {
	if len(expr) > 20 {
		return expr[:20] + "..."
	}
	return expr
}
