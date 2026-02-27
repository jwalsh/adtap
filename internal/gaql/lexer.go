package gaql

import (
	"strings"
	"unicode"
)

// Lexer tokenizes GAQL input.
type Lexer struct {
	input   string
	pos     int
	line    int
	column  int
	tokens  []Token
}

// NewLexer creates a new lexer for the given input.
func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  input,
		pos:    0,
		line:   1,
		column: 1,
	}
}

// Tokenize returns all tokens from the input.
func (l *Lexer) Tokenize() ([]Token, error) {
	for {
		tok := l.nextToken()
		l.tokens = append(l.tokens, tok)
		if tok.Type == TokenEOF {
			break
		}
		if tok.Type == TokenError {
			return l.tokens, &ParseError{
				Message: tok.Value,
				Line:    tok.Line,
				Column:  tok.Column,
			}
		}
	}
	return l.tokens, nil
}

func (l *Lexer) nextToken() Token {
	l.skipWhitespace()

	if l.pos >= len(l.input) {
		return Token{Type: TokenEOF, Line: l.line, Column: l.column}
	}

	ch := l.input[l.pos]
	startLine := l.line
	startCol := l.column

	// Single character tokens
	switch ch {
	case ',':
		l.advance()
		return Token{Type: TokenComma, Value: ",", Line: startLine, Column: startCol}
	case '(':
		l.advance()
		return Token{Type: TokenLParen, Value: "(", Line: startLine, Column: startCol}
	case ')':
		l.advance()
		return Token{Type: TokenRParen, Value: ")", Line: startLine, Column: startCol}
	case '.':
		l.advance()
		return Token{Type: TokenDot, Value: ".", Line: startLine, Column: startCol}
	case '=':
		l.advance()
		return Token{Type: TokenEq, Value: "=", Line: startLine, Column: startCol}
	case '!':
		if l.peek(1) == '=' {
			l.advance()
			l.advance()
			return Token{Type: TokenNeq, Value: "!=", Line: startLine, Column: startCol}
		}
		return Token{Type: TokenError, Value: "unexpected character '!'", Line: startLine, Column: startCol}
	case '>':
		if l.peek(1) == '=' {
			l.advance()
			l.advance()
			return Token{Type: TokenGte, Value: ">=", Line: startLine, Column: startCol}
		}
		l.advance()
		return Token{Type: TokenGt, Value: ">", Line: startLine, Column: startCol}
	case '<':
		if l.peek(1) == '=' {
			l.advance()
			l.advance()
			return Token{Type: TokenLte, Value: "<=", Line: startLine, Column: startCol}
		}
		l.advance()
		return Token{Type: TokenLt, Value: "<", Line: startLine, Column: startCol}
	case '\'', '"':
		return l.readString(ch)
	}

	// Numbers (including negative)
	if ch == '-' || isDigit(ch) {
		return l.readNumber()
	}

	// Identifiers and keywords
	if isLetter(ch) || ch == '_' {
		return l.readIdentOrKeyword()
	}

	l.advance()
	return Token{Type: TokenError, Value: "unexpected character '" + string(ch) + "'", Line: startLine, Column: startCol}
}

func (l *Lexer) readString(quote byte) Token {
	startLine := l.line
	startCol := l.column
	l.advance() // consume opening quote

	var sb strings.Builder
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == quote {
			l.advance() // consume closing quote
			return Token{Type: TokenString, Value: sb.String(), Line: startLine, Column: startCol}
		}
		if ch == '\\' && l.pos+1 < len(l.input) {
			l.advance()
			escaped := l.input[l.pos]
			switch escaped {
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			case '\\':
				sb.WriteByte('\\')
			case '\'':
				sb.WriteByte('\'')
			case '"':
				sb.WriteByte('"')
			default:
				sb.WriteByte(escaped)
			}
			l.advance()
			continue
		}
		sb.WriteByte(ch)
		l.advance()
	}

	return Token{Type: TokenError, Value: "unterminated string", Line: startLine, Column: startCol}
}

func (l *Lexer) readNumber() Token {
	startLine := l.line
	startCol := l.column
	startPos := l.pos

	// Handle negative sign
	if l.input[l.pos] == '-' {
		l.advance()
	}

	// Read integer part
	for l.pos < len(l.input) && isDigit(l.input[l.pos]) {
		l.advance()
	}

	// Read decimal part
	if l.pos < len(l.input) && l.input[l.pos] == '.' {
		l.advance()
		for l.pos < len(l.input) && isDigit(l.input[l.pos]) {
			l.advance()
		}
	}

	value := l.input[startPos:l.pos]
	return Token{Type: TokenNumber, Value: value, Line: startLine, Column: startCol}
}

func (l *Lexer) readIdentOrKeyword() Token {
	startLine := l.line
	startCol := l.column
	startPos := l.pos

	for l.pos < len(l.input) && (isLetter(l.input[l.pos]) || isDigit(l.input[l.pos]) || l.input[l.pos] == '_') {
		l.advance()
	}

	value := l.input[startPos:l.pos]
	upper := strings.ToUpper(value)

	// Check for ORDER BY (two-word keyword)
	if upper == "ORDER" {
		l.skipWhitespace()
		if l.pos+2 <= len(l.input) && strings.ToUpper(l.input[l.pos:l.pos+2]) == "BY" {
			l.advance()
			l.advance()
			return Token{Type: TokenOrderBy, Value: "ORDER BY", Line: startLine, Column: startCol}
		}
		return Token{Type: TokenIdent, Value: value, Line: startLine, Column: startCol}
	}

	// Check for date range keywords
	if _, ok := DateRangeKeywords[upper]; ok {
		return Token{Type: TokenDateRange, Value: upper, Line: startLine, Column: startCol}
	}

	// Check for other keywords
	if tokType, ok := Keywords[upper]; ok {
		return Token{Type: tokType, Value: upper, Line: startLine, Column: startCol}
	}

	return Token{Type: TokenIdent, Value: value, Line: startLine, Column: startCol}
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == ' ' || ch == '\t' || ch == '\r' {
			l.column++
			l.pos++
		} else if ch == '\n' {
			l.line++
			l.column = 1
			l.pos++
		} else {
			break
		}
	}
}

func (l *Lexer) advance() {
	if l.pos < len(l.input) {
		if l.input[l.pos] == '\n' {
			l.line++
			l.column = 1
		} else {
			l.column++
		}
		l.pos++
	}
}

func (l *Lexer) peek(offset int) byte {
	pos := l.pos + offset
	if pos >= len(l.input) {
		return 0
	}
	return l.input[pos]
}

func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch))
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}
