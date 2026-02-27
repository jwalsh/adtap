package gaql

// TokenType represents the type of a lexical token.
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenError

	// Keywords
	TokenSelect
	TokenFrom
	TokenWhere
	TokenOrderBy
	TokenLimit
	TokenParameters
	TokenAnd
	TokenOr
	TokenNot
	TokenAsc
	TokenDesc
	TokenIn
	TokenLike
	TokenContains
	TokenAny
	TokenAll
	TokenNone
	TokenIs
	TokenNull
	TokenDuring
	TokenBetween
	TokenRegexpMatch

	// Literals
	TokenIdent      // field names, resource names
	TokenString     // 'string' or "string"
	TokenNumber     // 123, 45.67, -123
	TokenDateRange  // TODAY, YESTERDAY, LAST_7_DAYS, etc.

	// Operators
	TokenEq    // =
	TokenNeq   // !=
	TokenGt    // >
	TokenGte   // >=
	TokenLt    // <
	TokenLte   // <=

	// Punctuation
	TokenComma      // ,
	TokenLParen     // (
	TokenRParen     // )
	TokenDot        // .
)

// Token represents a lexical token.
type Token struct {
	Type    TokenType
	Value   string
	Line    int
	Column  int
}

func (t TokenType) String() string {
	switch t {
	case TokenEOF:
		return "EOF"
	case TokenError:
		return "ERROR"
	case TokenSelect:
		return "SELECT"
	case TokenFrom:
		return "FROM"
	case TokenWhere:
		return "WHERE"
	case TokenOrderBy:
		return "ORDER BY"
	case TokenLimit:
		return "LIMIT"
	case TokenParameters:
		return "PARAMETERS"
	case TokenAnd:
		return "AND"
	case TokenOr:
		return "OR"
	case TokenNot:
		return "NOT"
	case TokenAsc:
		return "ASC"
	case TokenDesc:
		return "DESC"
	case TokenIn:
		return "IN"
	case TokenLike:
		return "LIKE"
	case TokenContains:
		return "CONTAINS"
	case TokenAny:
		return "ANY"
	case TokenAll:
		return "ALL"
	case TokenNone:
		return "NONE"
	case TokenIs:
		return "IS"
	case TokenNull:
		return "NULL"
	case TokenDuring:
		return "DURING"
	case TokenBetween:
		return "BETWEEN"
	case TokenRegexpMatch:
		return "REGEXP_MATCH"
	case TokenIdent:
		return "IDENT"
	case TokenString:
		return "STRING"
	case TokenNumber:
		return "NUMBER"
	case TokenDateRange:
		return "DATE_RANGE"
	case TokenEq:
		return "="
	case TokenNeq:
		return "!="
	case TokenGt:
		return ">"
	case TokenGte:
		return ">="
	case TokenLt:
		return "<"
	case TokenLte:
		return "<="
	case TokenComma:
		return ","
	case TokenLParen:
		return "("
	case TokenRParen:
		return ")"
	case TokenDot:
		return "."
	default:
		return "UNKNOWN"
	}
}

// Keywords maps keyword strings to token types.
var Keywords = map[string]TokenType{
	"SELECT":       TokenSelect,
	"FROM":         TokenFrom,
	"WHERE":        TokenWhere,
	"ORDER":        TokenIdent, // Handled specially with BY
	"BY":           TokenIdent, // Handled specially with ORDER
	"LIMIT":        TokenLimit,
	"PARAMETERS":   TokenParameters,
	"AND":          TokenAnd,
	"OR":           TokenOr,
	"NOT":          TokenNot,
	"ASC":          TokenAsc,
	"DESC":         TokenDesc,
	"IN":           TokenIn,
	"LIKE":         TokenLike,
	"CONTAINS":     TokenContains,
	"ANY":          TokenAny,
	"ALL":          TokenAll,
	"NONE":         TokenNone,
	"IS":           TokenIs,
	"NULL":         TokenNull,
	"DURING":       TokenDuring,
	"BETWEEN":      TokenBetween,
	"REGEXP_MATCH": TokenRegexpMatch,
}
