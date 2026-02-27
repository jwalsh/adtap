package gaql

import (
	"strconv"
	"strings"
)

// Parser parses GAQL queries into an AST.
type Parser struct {
	tokens []Token
	pos    int
}

// Parse parses a GAQL query string and returns the AST.
func Parse(input string) (*Query, error) {
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, err
	}

	p := &Parser{tokens: tokens, pos: 0}
	return p.parseQuery()
}

func (p *Parser) parseQuery() (*Query, error) {
	query := &Query{
		Parameters: make(map[string]string),
	}

	// Parse SELECT clause (required)
	if !p.match(TokenSelect) {
		return nil, p.error("expected SELECT clause")
	}

	fields, err := p.parseFieldList()
	if err != nil {
		return nil, err
	}
	query.Select = fields

	// Parse FROM clause (required)
	if !p.match(TokenFrom) {
		return nil, p.error("expected FROM clause")
	}

	if !p.check(TokenIdent) {
		return nil, p.error("expected resource name after FROM")
	}
	query.From = p.current().Value
	p.advance()

	// Parse optional WHERE clause
	if p.match(TokenWhere) {
		conditions, err := p.parseConditions()
		if err != nil {
			return nil, err
		}
		query.Where = conditions
	}

	// Parse optional ORDER BY clause
	if p.match(TokenOrderBy) {
		orderings, err := p.parseOrderings()
		if err != nil {
			return nil, err
		}
		query.OrderBy = orderings
	}

	// Parse optional LIMIT clause
	if p.match(TokenLimit) {
		if !p.check(TokenNumber) {
			return nil, p.error("expected number after LIMIT")
		}
		limit, err := strconv.Atoi(p.current().Value)
		if err != nil {
			return nil, p.error("invalid LIMIT value: " + p.current().Value)
		}
		if limit <= 0 {
			return nil, p.error("LIMIT must be a positive integer")
		}
		query.Limit = limit
		p.advance()
	}

	// Parse optional PARAMETERS clause
	if p.match(TokenParameters) {
		params, err := p.parseParameters()
		if err != nil {
			return nil, err
		}
		query.Parameters = params
	}

	// Should be at EOF
	if !p.check(TokenEOF) {
		return nil, p.error("unexpected token: " + p.current().Value)
	}

	return query, nil
}

func (p *Parser) parseFieldList() ([]Field, error) {
	var fields []Field

	for {
		field, err := p.parseField()
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)

		if !p.match(TokenComma) {
			break
		}
	}

	if len(fields) == 0 {
		return nil, p.error("SELECT must contain at least one field")
	}

	return fields, nil
}

func (p *Parser) parseField() (Field, error) {
	var parts []string

	if !p.check(TokenIdent) {
		return Field{}, p.error("expected field name")
	}
	parts = append(parts, p.current().Value)
	p.advance()

	// Handle dotted field names (e.g., campaign.id, metrics.clicks)
	for p.match(TokenDot) {
		if !p.check(TokenIdent) {
			return Field{}, p.error("expected field name after '.'")
		}
		parts = append(parts, p.current().Value)
		p.advance()
	}

	return Field{Name: strings.Join(parts, ".")}, nil
}

func (p *Parser) parseConditions() ([]Condition, error) {
	var conditions []Condition

	for {
		cond, err := p.parseCondition()
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, cond)

		if !p.match(TokenAnd) {
			break
		}
	}

	return conditions, nil
}

func (p *Parser) parseCondition() (Condition, error) {
	cond := Condition{}

	// Parse field name
	field, err := p.parseField()
	if err != nil {
		return cond, err
	}
	cond.Field = field.Name

	// Parse operator
	op, err := p.parseOperator()
	if err != nil {
		return cond, err
	}
	cond.Operator = op

	// Parse value (not needed for IS NULL, IS NOT NULL)
	if op == OpIsNull || op == OpIsNotNull {
		cond.Value = Value{Type: ValueNull}
		return cond, nil
	}

	value, err := p.parseValue(op)
	if err != nil {
		return cond, err
	}
	cond.Value = value

	return cond, nil
}

func (p *Parser) parseOperator() (Operator, error) {
	tok := p.current()

	switch tok.Type {
	case TokenEq:
		p.advance()
		return OpEq, nil
	case TokenNeq:
		p.advance()
		return OpNeq, nil
	case TokenGt:
		p.advance()
		return OpGt, nil
	case TokenGte:
		p.advance()
		return OpGte, nil
	case TokenLt:
		p.advance()
		return OpLt, nil
	case TokenLte:
		p.advance()
		return OpLte, nil
	case TokenIn:
		p.advance()
		return OpIn, nil
	case TokenNot:
		p.advance()
		if p.match(TokenIn) {
			return OpNotIn, nil
		}
		if p.match(TokenLike) {
			return OpNotLike, nil
		}
		if p.match(TokenRegexpMatch) {
			return OpNotRegexpMatch, nil
		}
		return 0, p.error("expected IN, LIKE, or REGEXP_MATCH after NOT")
	case TokenLike:
		p.advance()
		return OpLike, nil
	case TokenContains:
		p.advance()
		if p.match(TokenAny) {
			return OpContainsAny, nil
		}
		if p.match(TokenAll) {
			return OpContainsAll, nil
		}
		if p.match(TokenNone) {
			return OpContainsNone, nil
		}
		return 0, p.error("expected ANY, ALL, or NONE after CONTAINS")
	case TokenIs:
		p.advance()
		if p.match(TokenNot) {
			if !p.match(TokenNull) {
				return 0, p.error("expected NULL after IS NOT")
			}
			return OpIsNotNull, nil
		}
		if !p.match(TokenNull) {
			return 0, p.error("expected NULL or NOT NULL after IS")
		}
		return OpIsNull, nil
	case TokenDuring:
		p.advance()
		return OpDuring, nil
	case TokenBetween:
		p.advance()
		return OpBetween, nil
	case TokenRegexpMatch:
		p.advance()
		return OpRegexpMatch, nil
	default:
		return 0, p.error("expected operator, got " + tok.Type.String())
	}
}

func (p *Parser) parseValue(op Operator) (Value, error) {
	tok := p.current()

	// Handle DURING keyword values
	if op == OpDuring {
		if !p.check(TokenDateRange) {
			return Value{}, p.error("expected date range keyword after DURING")
		}
		dr, ok := DateRangeKeywords[tok.Value]
		if !ok {
			return Value{}, p.error("unknown date range: " + tok.Value)
		}
		p.advance()
		return Value{Type: ValueDateRange, DateRange: dr}, nil
	}

	// Handle BETWEEN
	if op == OpBetween {
		start, err := p.parseSimpleValue()
		if err != nil {
			return Value{}, err
		}
		if !p.match(TokenAnd) {
			return Value{}, p.error("expected AND in BETWEEN clause")
		}
		end, err := p.parseSimpleValue()
		if err != nil {
			return Value{}, err
		}
		return Value{
			Type: ValueList,
			List: []string{start, end},
		}, nil
	}

	// Handle IN/NOT IN lists
	if op == OpIn || op == OpNotIn || op == OpContainsAny || op == OpContainsAll || op == OpContainsNone {
		return p.parseList()
	}

	// Handle simple values
	switch tok.Type {
	case TokenString:
		p.advance()
		return Value{Type: ValueString, Str: tok.Value}, nil
	case TokenNumber:
		num, err := strconv.ParseFloat(tok.Value, 64)
		if err != nil {
			return Value{}, p.error("invalid number: " + tok.Value)
		}
		p.advance()
		return Value{Type: ValueNumber, Number: num}, nil
	case TokenIdent:
		// Could be an enum value without quotes
		p.advance()
		return Value{Type: ValueString, Str: tok.Value}, nil
	default:
		return Value{}, p.error("expected value, got " + tok.Type.String())
	}
}

func (p *Parser) parseSimpleValue() (string, error) {
	tok := p.current()
	switch tok.Type {
	case TokenString:
		p.advance()
		return tok.Value, nil
	case TokenNumber:
		p.advance()
		return tok.Value, nil
	case TokenIdent:
		p.advance()
		return tok.Value, nil
	default:
		return "", p.error("expected value, got " + tok.Type.String())
	}
}

func (p *Parser) parseList() (Value, error) {
	if !p.match(TokenLParen) {
		return Value{}, p.error("expected '(' before list")
	}

	var items []string
	for {
		val, err := p.parseSimpleValue()
		if err != nil {
			return Value{}, err
		}
		items = append(items, val)

		if !p.match(TokenComma) {
			break
		}
	}

	if !p.match(TokenRParen) {
		return Value{}, p.error("expected ')' after list")
	}

	return Value{Type: ValueList, List: items}, nil
}

func (p *Parser) parseOrderings() ([]Ordering, error) {
	var orderings []Ordering

	for {
		field, err := p.parseField()
		if err != nil {
			return nil, err
		}

		dir := Asc
		if p.match(TokenDesc) {
			dir = Desc
		} else if p.match(TokenAsc) {
			dir = Asc
		}

		orderings = append(orderings, Ordering{Field: field.Name, Direction: dir})

		if !p.match(TokenComma) {
			break
		}
	}

	return orderings, nil
}

func (p *Parser) parseParameters() (map[string]string, error) {
	params := make(map[string]string)

	for {
		if !p.check(TokenIdent) {
			return nil, p.error("expected parameter name")
		}
		name := p.current().Value
		p.advance()

		if !p.match(TokenEq) {
			return nil, p.error("expected '=' after parameter name")
		}

		val, err := p.parseSimpleValue()
		if err != nil {
			return nil, err
		}
		params[name] = val

		if !p.match(TokenComma) {
			break
		}
	}

	return params, nil
}

// Helper methods

func (p *Parser) current() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() {
	if p.pos < len(p.tokens) {
		p.pos++
	}
}

func (p *Parser) check(t TokenType) bool {
	return p.current().Type == t
}

func (p *Parser) match(t TokenType) bool {
	if p.check(t) {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) error(msg string) error {
	tok := p.current()
	return &ParseError{
		Message: msg,
		Line:    tok.Line,
		Column:  tok.Column,
	}
}
