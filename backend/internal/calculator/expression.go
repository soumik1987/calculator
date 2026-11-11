package calculator

import "fmt"

// SyntaxError reports why an expression could not be read. Its message is a full
// sentence meant for users, e.g. `Unexpected ")" at position 4.`.
type SyntaxError struct {
	msg string
}

func syntaxErrorf(format string, args ...any) *SyntaxError {
	return &SyntaxError{msg: fmt.Sprintf(format, args...)}
}

func (e *SyntaxError) Error() string {
	return e.msg
}

// Evaluate reads an expression such as "(986 × (4 + 8)) + √16" and returns its value.
func Evaluate(expression string) (float64, error) {
	tree, err := parse(expression)
	if err != nil {
		return 0, err
	}
	return tree.eval()
}

// node is one part of a parsed expression: a number, or an operator applied to other nodes.
type node interface {
	eval() (float64, error)
}

type number float64

func (n number) eval() (float64, error) {
	return float64(n), nil
}

type unary struct {
	fn      unaryFunc
	operand node
}

func (u unary) eval() (float64, error) {
	a, err := u.operand.eval()
	if err != nil {
		return 0, err
	}
	return finite(u.fn(a))
}

type binary struct {
	fn          binaryFunc
	left, right node
}

func (b binary) eval() (float64, error) {
	l, err := b.left.eval()
	if err != nil {
		return 0, err
	}
	r, err := b.right.eval()
	if err != nil {
		return 0, err
	}
	return finite(b.fn(l, r))
}
