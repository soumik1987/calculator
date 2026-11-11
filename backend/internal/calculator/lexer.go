package calculator

import (
	"strconv"
	"strings"
	"unicode"
)

type tokenKind int

const (
	tokenEnd tokenKind = iota
	tokenNumber
	tokenPlus
	tokenMinus
	tokenTimes
	tokenDivide
	tokenPower
	tokenPercent
	tokenFunction
	tokenOpen
	tokenClose
)

type token struct {
	kind  tokenKind
	text  string
	pos   int
	value float64
	fn    unaryFunc
}

// operators maps symbols and words to tokens. Words are matched case-insensitively.
var operators = map[string]tokenKind{
	"+": tokenPlus,
	"-": tokenMinus,
	"−": tokenMinus,
	"*": tokenTimes,
	"×": tokenTimes,
	"x": tokenTimes,
	"/": tokenDivide,
	"÷": tokenDivide,
	"^": tokenPower,
	"%": tokenPercent,
	"(": tokenOpen,
	")": tokenClose,
}

// functions apply to the number or bracket that follows them.
var functions = map[string]unaryFunc{
	"√":    squareRoot,
	"sqrt": squareRoot,
	"sin":  sine,
	"cos":  cosine,
	"log":  logarithm,
	"exp":  exponential,
}

// tokenize splits input into tokens. Positions are 1-based character offsets.
func tokenize(input []rune) ([]token, error) {
	var tokens []token
	for i := 0; i < len(input); {
		r, pos := input[i], i+1
		switch {
		case unicode.IsSpace(r):
			i++
		case isDigit(r) || r == '.':
			end := scanNumber(input, i)
			text := string(input[i:end])
			value, err := strconv.ParseFloat(text, 64)
			if err != nil {
				return nil, syntaxErrorf("Invalid number %q at position %d.", text, pos)
			}
			tokens = append(tokens, token{kind: tokenNumber, text: text, pos: pos, value: value})
			i = end
		case unicode.IsLetter(r):
			end := i
			for end < len(input) && unicode.IsLetter(input[end]) {
				end++
			}
			text := string(input[i:end])
			t, ok := lookup(text, pos)
			if !ok {
				return nil, syntaxErrorf("Unknown word %q at position %d.", text, pos)
			}
			tokens = append(tokens, t)
			i = end
		default:
			text := string(r)
			t, ok := lookup(text, pos)
			if !ok {
				return nil, syntaxErrorf("Unexpected character %q at position %d.", text, pos)
			}
			tokens = append(tokens, t)
			i++
		}
	}
	return append(tokens, token{kind: tokenEnd, pos: len(input) + 1}), nil
}

func lookup(text string, pos int) (token, bool) {
	name := strings.ToLower(text)
	if fn, ok := functions[name]; ok {
		return token{kind: tokenFunction, text: text, pos: pos, fn: fn}, true
	}
	kind, ok := operators[name]
	return token{kind: kind, text: text, pos: pos}, ok
}

// scanNumber returns the end of the number starting at i: digits and dots, then an
// optional exponent such as e5 or E-3.
func scanNumber(input []rune, i int) int {
	for i < len(input) && (isDigit(input[i]) || input[i] == '.') {
		i++
	}
	if i < len(input) && (input[i] == 'e' || input[i] == 'E') {
		j := i + 1
		if j < len(input) && (input[j] == '+' || input[j] == '-') {
			j++
		}
		if j < len(input) && isDigit(input[j]) {
			for j < len(input) && isDigit(input[j]) {
				j++
			}
			return j
		}
	}
	return i
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}
