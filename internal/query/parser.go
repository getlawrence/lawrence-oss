// Copyright (c) 2024 Lawrence OSS Contributors
// SPDX-License-Identifier: Apache-2.0

package query

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Parser parses Lawrence QL queries into AST
type Parser struct {
	input  string
	pos    int
	tokens []Token
}

// Token represents a lexical token
type Token struct {
	Type  TokenType
	Value string
	Pos   int
}

// TokenType represents the type of token
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdentifier
	TokenString
	TokenNumber
	TokenLBrace   // {
	TokenRBrace   // }
	TokenLBracket // [
	TokenRBracket // ]
	TokenLParen   // (
	TokenRParen   // )
	TokenComma    // ,
	TokenEqual    // =
	TokenNotEqual // !=
	TokenRegex    // =~
	TokenNotRegex // !~
	TokenPlus     // +
	TokenMinus    // -
	TokenMultiply // *
	TokenDivide   // /
	TokenLT       // <
	TokenGT       // >
	TokenLTE      // <=
	TokenGTE      // >=
	TokenAnd      // and
	TokenOr       // or
	TokenBy       // by
	TokenPipe     // |
)

// NewParser creates a new parser for the given input
func NewParser(input string) *Parser {
	return &Parser{
		input:  input,
		pos:    0,
		tokens: []Token{},
	}
}

// Parse parses the input and returns the AST
func (p *Parser) Parse() (Query, error) {
	// Tokenize
	if err := p.tokenize(); err != nil {
		return nil, err
	}

	// Reset position for parsing
	p.pos = 0

	// Parse query
	return p.parseQuery()
}

// tokenize converts input string into tokens
func (p *Parser) tokenize() error {
	input := p.input
	pos := 0

	for pos < len(input) {
		// Skip whitespace
		if input[pos] == ' ' || input[pos] == '\t' || input[pos] == '\n' || input[pos] == '\r' {
			pos++
			continue
		}

		// Multi-character operators
		if pos+1 < len(input) {
			twoChar := input[pos : pos+2]
			switch twoChar {
			case "!=":
				p.tokens = append(p.tokens, Token{Type: TokenNotEqual, Value: "!=", Pos: pos})
				pos += 2
				continue
			case "=~":
				p.tokens = append(p.tokens, Token{Type: TokenRegex, Value: "=~", Pos: pos})
				pos += 2
				continue
			case "!~":
				p.tokens = append(p.tokens, Token{Type: TokenNotRegex, Value: "!~", Pos: pos})
				pos += 2
				continue
			case "<=":
				p.tokens = append(p.tokens, Token{Type: TokenLTE, Value: "<=", Pos: pos})
				pos += 2
				continue
			case ">=":
				p.tokens = append(p.tokens, Token{Type: TokenGTE, Value: ">=", Pos: pos})
				pos += 2
				continue
			case "==":
				p.tokens = append(p.tokens, Token{Type: TokenEqual, Value: "==", Pos: pos})
				pos += 2
				continue
			}
		}

		// Single character tokens
		switch input[pos] {
		case '{':
			p.tokens = append(p.tokens, Token{Type: TokenLBrace, Value: "{", Pos: pos})
			pos++
		case '}':
			p.tokens = append(p.tokens, Token{Type: TokenRBrace, Value: "}", Pos: pos})
			pos++
		case '[':
			p.tokens = append(p.tokens, Token{Type: TokenLBracket, Value: "[", Pos: pos})
			pos++
		case ']':
			p.tokens = append(p.tokens, Token{Type: TokenRBracket, Value: "]", Pos: pos})
			pos++
		case '(':
			p.tokens = append(p.tokens, Token{Type: TokenLParen, Value: "(", Pos: pos})
			pos++
		case ')':
			p.tokens = append(p.tokens, Token{Type: TokenRParen, Value: ")", Pos: pos})
			pos++
		case ',':
			p.tokens = append(p.tokens, Token{Type: TokenComma, Value: ",", Pos: pos})
			pos++
		case '=':
			p.tokens = append(p.tokens, Token{Type: TokenEqual, Value: "=", Pos: pos})
			pos++
		case '+':
			p.tokens = append(p.tokens, Token{Type: TokenPlus, Value: "+", Pos: pos})
			pos++
		case '-':
			p.tokens = append(p.tokens, Token{Type: TokenMinus, Value: "-", Pos: pos})
			pos++
		case '*':
			p.tokens = append(p.tokens, Token{Type: TokenMultiply, Value: "*", Pos: pos})
			pos++
		case '/':
			p.tokens = append(p.tokens, Token{Type: TokenDivide, Value: "/", Pos: pos})
			pos++
		case '<':
			p.tokens = append(p.tokens, Token{Type: TokenLT, Value: "<", Pos: pos})
			pos++
		case '>':
			p.tokens = append(p.tokens, Token{Type: TokenGT, Value: ">", Pos: pos})
			pos++
		case '|':
			p.tokens = append(p.tokens, Token{Type: TokenPipe, Value: "|", Pos: pos})
			pos++
		case '"':
			// String literal
			start := pos
			pos++
			for pos < len(input) && input[pos] != '"' {
				if input[pos] == '\\' && pos+1 < len(input) {
					pos += 2
				} else {
					pos++
				}
			}
			if pos >= len(input) {
				return fmt.Errorf("unterminated string at position %d", start)
			}
			pos++ // closing quote
			p.tokens = append(p.tokens, Token{Type: TokenString, Value: input[start+1 : pos-1], Pos: start})
		default:
			// Identifier or number
			if isAlpha(input[pos]) {
				start := pos
				for pos < len(input) && (isAlphaNumeric(input[pos]) || input[pos] == '_') {
					pos++
				}
				value := input[start:pos]
				tokenType := TokenIdentifier

				// Check for keywords
				switch value {
				case "and":
					tokenType = TokenAnd
				case "or":
					tokenType = TokenOr
				case "by":
					tokenType = TokenBy
				}

				p.tokens = append(p.tokens, Token{Type: tokenType, Value: value, Pos: start})
			} else if isDigit(input[pos]) {
				start := pos
				for pos < len(input) && (isDigit(input[pos]) || input[pos] == '.') {
					pos++
				}
				// Check if this is followed by a duration unit (s, m, h, d)
				if pos < len(input) && (input[pos] == 's' || input[pos] == 'm' || input[pos] == 'h' || input[pos] == 'd') {
					pos++
				}
				p.tokens = append(p.tokens, Token{Type: TokenNumber, Value: input[start:pos], Pos: start})
			} else {
				return fmt.Errorf("unexpected character '%c' at position %d", input[pos], pos)
			}
		}
	}

	p.tokens = append(p.tokens, Token{Type: TokenEOF, Value: "", Pos: len(input)})
	return nil
}

// parseQuery parses a query expression
func (p *Parser) parseQuery() (Query, error) {
	return p.parseBinaryOp()
}

// parseBinaryOp parses binary operations
func (p *Parser) parseBinaryOp() (Query, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	for {
		token := p.peek()
		if token.Type == TokenEOF {
			break
		}

		var op BinaryOperator
		switch token.Type {
		case TokenPlus:
			op = BinaryOpAdd
		case TokenMinus:
			op = BinaryOpSubtract
		case TokenMultiply:
			op = BinaryOpMultiply
		case TokenDivide:
			op = BinaryOpDivide
		case TokenAnd:
			op = BinaryOpAnd
		case TokenOr:
			op = BinaryOpOr
		case TokenLT:
			op = BinaryOpLT
		case TokenGT:
			op = BinaryOpGT
		case TokenLTE:
			op = BinaryOpLTE
		case TokenGTE:
			op = BinaryOpGTE
		default:
			return left, nil
		}

		p.consume()
		right, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}

		left = &BinaryOp{
			Left:     left,
			Operator: op,
			Right:    right,
		}
	}

	return left, nil
}

// parsePrimary parses primary expressions
func (p *Parser) parsePrimary() (Query, error) {
	token := p.peek()

	// Function call or aggregation
	if token.Type == TokenIdentifier {
		// Check if it's a telemetry type or function
		if token.Value == "metrics" || token.Value == "logs" || token.Value == "traces" {
			return p.parseTelemetryQuery()
		}
		// Otherwise it's a function call
		return p.parseFunctionCall()
	}

	// Parenthesized expression
	if token.Type == TokenLParen {
		p.consume()
		query, err := p.parseQuery()
		if err != nil {
			return nil, err
		}
		if p.peek().Type != TokenRParen {
			return nil, fmt.Errorf("expected ')' at position %d", p.peek().Pos)
		}
		p.consume()
		return query, nil
	}

	return nil, fmt.Errorf("unexpected token '%s' at position %d", token.Value, token.Pos)
}

// parseTelemetryQuery parses a telemetry query: type{selectors}[duration]
func (p *Parser) parseTelemetryQuery() (*TelemetryQuery, error) {
	// Parse type
	typeToken := p.consume()
	var telemetryType TelemetryType
	switch typeToken.Value {
	case "metrics":
		telemetryType = TelemetryTypeMetrics
	case "logs":
		telemetryType = TelemetryTypeLogs
	case "traces":
		telemetryType = TelemetryTypeTraces
	default:
		return nil, fmt.Errorf("invalid telemetry type '%s' at position %d", typeToken.Value, typeToken.Pos)
	}

	// Parse selectors: {label=value, ...}
	if p.peek().Type != TokenLBrace {
		return nil, fmt.Errorf("expected '{' at position %d", p.peek().Pos)
	}
	p.consume()

	selectors := make(map[string]*Selector)
	for p.peek().Type != TokenRBrace {
		selector, err := p.parseSelector()
		if err != nil {
			return nil, err
		}
		selectors[selector.Label] = selector

		if p.peek().Type == TokenComma {
			p.consume()
		} else if p.peek().Type != TokenRBrace {
			return nil, fmt.Errorf("expected ',' or '}' at position %d", p.peek().Pos)
		}
	}
	p.consume() // consume '}'

	// Parse duration: [5m]
	var duration time.Duration
	if p.peek().Type == TokenLBracket {
		p.consume()
		durationToken := p.consume()
		var err error
		duration, err = parseDuration(durationToken.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid duration '%s' at position %d: %v", durationToken.Value, durationToken.Pos, err)
		}
		if p.peek().Type != TokenRBracket {
			return nil, fmt.Errorf("expected ']' at position %d", p.peek().Pos)
		}
		p.consume()
	}

	return &TelemetryQuery{
		Type:      telemetryType,
		Selectors: selectors,
		Duration:  duration,
		Limit:     1000, // default
	}, nil
}

// parseSelector parses a selector: label=value or label=~"regex"
func (p *Parser) parseSelector() (*Selector, error) {
	labelToken := p.consume()
	if labelToken.Type != TokenIdentifier {
		return nil, fmt.Errorf("expected label identifier at position %d", labelToken.Pos)
	}

	opToken := p.consume()
	var op SelectorOperator
	switch opToken.Type {
	case TokenEqual:
		op = SelectorOpEqual
	case TokenNotEqual:
		op = SelectorOpNotEqual
	case TokenRegex:
		op = SelectorOpRegex
	case TokenNotRegex:
		op = SelectorOpNotRegex
	default:
		return nil, fmt.Errorf("expected selector operator at position %d", opToken.Pos)
	}

	valueToken := p.consume()
	if valueToken.Type != TokenString && valueToken.Type != TokenIdentifier {
		return nil, fmt.Errorf("expected value at position %d", valueToken.Pos)
	}

	return &Selector{
		Label:    labelToken.Value,
		Operator: op,
		Value:    valueToken.Value,
	}, nil
}

// parseFunctionCall parses a function call: func(query)
func (p *Parser) parseFunctionCall() (Query, error) {
	nameToken := p.consume()
	if nameToken.Type != TokenIdentifier {
		return nil, fmt.Errorf("expected function name at position %d", nameToken.Pos)
	}

	if p.peek().Type != TokenLParen {
		return nil, fmt.Errorf("expected '(' at position %d", p.peek().Pos)
	}
	p.consume()

	// Parse arguments
	args := []Query{}
	for p.peek().Type != TokenRParen {
		arg, err := p.parseQuery()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		if p.peek().Type == TokenComma {
			p.consume()
		} else if p.peek().Type != TokenRParen {
			return nil, fmt.Errorf("expected ',' or ')' at position %d", p.peek().Pos)
		}
	}
	p.consume() // consume ')'

	// Check for aggregation with 'by' clause: sum(...) by (label1, label2)
	if p.peek().Type == TokenBy {
		p.consume()
		if p.peek().Type != TokenLParen {
			return nil, fmt.Errorf("expected '(' after 'by' at position %d", p.peek().Pos)
		}
		p.consume()

		byLabels := []string{}
		for p.peek().Type != TokenRParen {
			labelToken := p.consume()
			if labelToken.Type != TokenIdentifier {
				return nil, fmt.Errorf("expected label identifier at position %d", labelToken.Pos)
			}
			byLabels = append(byLabels, labelToken.Value)

			if p.peek().Type == TokenComma {
				p.consume()
			} else if p.peek().Type != TokenRParen {
				return nil, fmt.Errorf("expected ',' or ')' at position %d", p.peek().Pos)
			}
		}
		p.consume() // consume ')'

		return &Aggregation{
			Function: nameToken.Value,
			Query:    args[0],
			By:       byLabels,
		}, nil
	}

	return &FunctionCall{
		Name: nameToken.Value,
		Args: args,
	}, nil
}

// peek returns the current token without consuming it
func (p *Parser) peek() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokenEOF, Value: "", Pos: len(p.input)}
	}
	return p.tokens[p.pos]
}

// consume returns the current token and advances
func (p *Parser) consume() Token {
	token := p.peek()
	p.pos++
	return token
}

// parseDuration parses a duration string (e.g., "5m", "1h", "7d")
func parseDuration(s string) (time.Duration, error) {
	if len(s) < 2 {
		return 0, fmt.Errorf("invalid duration format")
	}

	valueStr := s[:len(s)-1]
	unit := s[len(s)-1]

	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return 0, err
	}

	switch unit {
	case 's':
		return time.Duration(value) * time.Second, nil
	case 'm':
		return time.Duration(value) * time.Minute, nil
	case 'h':
		return time.Duration(value) * time.Hour, nil
	case 'd':
		return time.Duration(value) * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("invalid duration unit '%c'", unit)
	}
}

// isAlpha checks if a character is alphabetic
func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// isDigit checks if a character is a digit
func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// isAlphaNumeric checks if a character is alphanumeric
func isAlphaNumeric(c byte) bool {
	return isAlpha(c) || isDigit(c)
}

// ValidateQuery validates a Lawrence QL query string
func ValidateQuery(queryStr string) error {
	parser := NewParser(queryStr)
	_, err := parser.Parse()
	return err
}

// GetQuerySuggestions returns auto-completion suggestions for a query
func GetQuerySuggestions(queryStr string, cursorPos int) []string {
	// Basic suggestions - can be enhanced later
	suggestions := []string{
		"metrics{",
		"logs{",
		"traces{",
		"sum(",
		"avg(",
		"min(",
		"max(",
		"count(",
		"rate(",
		"increase(",
	}

	// Filter based on what's already typed
	if queryStr == "" {
		return suggestions
	}

	filtered := []string{}
	prefix := strings.ToLower(queryStr[:cursorPos])
	for _, s := range suggestions {
		if strings.HasPrefix(strings.ToLower(s), prefix) {
			filtered = append(filtered, s)
		}
	}

	return filtered
}

// ValidateSelector validates a selector value based on operator
func (s *Selector) Validate() error {
	if s.Operator == SelectorOpRegex || s.Operator == SelectorOpNotRegex {
		_, err := regexp.Compile(s.Value)
		if err != nil {
			return fmt.Errorf("invalid regex pattern: %v", err)
		}
	}
	return nil
}
