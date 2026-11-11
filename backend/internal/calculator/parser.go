package calculator

import "strings"

// maxExpressionLength also bounds how deep the parser can recurse.
const maxExpressionLength = 500

// parse turns an expression into a tree. Grammar, lowest precedence first:
//
//	expression = term { ("+" | "-") term }
//	term       = unary { ("*" | "/") unary }
//	unary      = ("+" | "-") unary | power
//	power      = postfix [ "^" unary ]
//	postfix    = primary { "%" }
//	primary    = number | "(" expression ")" | function primary
//
// function is √, sqrt, sin, cos, log or exp. So -2^2 is -(2^2), 2^3^2 is 2^(3^2), and
// √16^2 is (√16)^2.
func parse(expression string) (node, error) {
	input := []rune(expression)
	if len(input) > maxExpressionLength {
		return nil, syntaxErrorf("The expression is longer than %d characters.", maxExpressionLength)
	}
	if strings.TrimSpace(expression) == "" {
		return nil, syntaxErrorf("Enter an expression.")
	}

	tokens, err := tokenize(input)
	if err != nil {
		return nil, err
	}
	p := &parser{tokens: tokens}
	tree, err := p.expression()
	if err != nil {
		return nil, err
	}
	if t := p.peek(); t.kind != tokenEnd {
		return nil, unexpected(t)
	}
	return tree, nil
}

type parser struct {
	tokens []token
	next   int
}

func (p *parser) peek() token {
	return p.tokens[p.next]
}

func (p *parser) advance() {
	p.next++
}

func (p *parser) expression() (node, error) {
	return p.binaryChain(p.term, map[tokenKind]binaryFunc{tokenPlus: add, tokenMinus: subtract})
}

func (p *parser) term() (node, error) {
	return p.binaryChain(p.unary, map[tokenKind]binaryFunc{tokenTimes: multiply, tokenDivide: divide})
}

// binaryChain parses operand { operator operand }, grouping from the left.
func (p *parser) binaryChain(operand func() (node, error), operators map[tokenKind]binaryFunc) (node, error) {
	left, err := operand()
	if err != nil {
		return nil, err
	}
	for {
		fn, ok := operators[p.peek().kind]
		if !ok {
			return left, nil
		}
		p.advance()
		right, err := operand()
		if err != nil {
			return nil, err
		}
		left = binary{fn: fn, left: left, right: right}
	}
}

func (p *parser) unary() (node, error) {
	kind := p.peek().kind
	if kind != tokenPlus && kind != tokenMinus {
		return p.power()
	}
	p.advance()
	operand, err := p.unary()
	if err != nil {
		return nil, err
	}
	if kind == tokenMinus {
		return unary{fn: negate, operand: operand}, nil
	}
	return operand, nil
}

func (p *parser) power() (node, error) {
	base, err := p.postfix()
	if err != nil {
		return nil, err
	}
	if p.peek().kind != tokenPower {
		return base, nil
	}
	p.advance()
	exponent, err := p.unary()
	if err != nil {
		return nil, err
	}
	return binary{fn: power, left: base, right: exponent}, nil
}

func (p *parser) postfix() (node, error) {
	n, err := p.primary()
	if err != nil {
		return nil, err
	}
	for p.peek().kind == tokenPercent {
		p.advance()
		n = unary{fn: percent, operand: n}
	}
	return n, nil
}

func (p *parser) primary() (node, error) {
	t := p.peek()
	switch t.kind {
	case tokenNumber:
		p.advance()
		return number(t.value), nil
	case tokenOpen, tokenFunction:
		p.advance()
		if t.kind == tokenFunction {
			operand, err := p.primary()
			if err != nil {
				return nil, err
			}
			return unary{fn: t.fn, operand: operand}, nil
		}
		inner, err := p.expression()
		if err != nil {
			return nil, err
		}
		if p.peek().kind != tokenClose {
			return nil, syntaxErrorf(`Missing ")" for the "(" at position %d.`, t.pos)
		}
		p.advance()
		return inner, nil
	case tokenEnd:
		return nil, syntaxErrorf("The expression is incomplete.")
	default:
		return nil, unexpected(t)
	}
}

func unexpected(t token) error {
	return syntaxErrorf("Unexpected %q at position %d.", t.text, t.pos)
}
