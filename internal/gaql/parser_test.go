package gaql

import (
	"testing"
)

func TestParseBasicQuery(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(*Query) error
	}{
		{
			name:  "simple select",
			input: "SELECT campaign.id FROM campaign",
			check: func(q *Query) error {
				if len(q.Select) != 1 {
					t.Errorf("expected 1 field, got %d", len(q.Select))
				}
				if q.Select[0].Name != "campaign.id" {
					t.Errorf("expected campaign.id, got %s", q.Select[0].Name)
				}
				if q.From != "campaign" {
					t.Errorf("expected campaign, got %s", q.From)
				}
				return nil
			},
		},
		{
			name:  "multiple fields",
			input: "SELECT campaign.id, campaign.name, campaign.status FROM campaign",
			check: func(q *Query) error {
				if len(q.Select) != 3 {
					t.Errorf("expected 3 fields, got %d", len(q.Select))
				}
				return nil
			},
		},
		{
			name:  "with where clause",
			input: "SELECT campaign.id FROM campaign WHERE campaign.status = 'ENABLED'",
			check: func(q *Query) error {
				if len(q.Where) != 1 {
					t.Errorf("expected 1 condition, got %d", len(q.Where))
				}
				if q.Where[0].Field != "campaign.status" {
					t.Errorf("expected campaign.status, got %s", q.Where[0].Field)
				}
				if q.Where[0].Operator != OpEq {
					t.Errorf("expected =, got %s", q.Where[0].Operator)
				}
				if q.Where[0].Value.Str != "ENABLED" {
					t.Errorf("expected ENABLED, got %s", q.Where[0].Value.Str)
				}
				return nil
			},
		},
		{
			name:  "with limit",
			input: "SELECT campaign.id FROM campaign LIMIT 10",
			check: func(q *Query) error {
				if q.Limit != 10 {
					t.Errorf("expected limit 10, got %d", q.Limit)
				}
				return nil
			},
		},
		{
			name:  "with order by",
			input: "SELECT campaign.id, metrics.clicks FROM campaign ORDER BY metrics.clicks DESC",
			check: func(q *Query) error {
				if len(q.OrderBy) != 1 {
					t.Errorf("expected 1 ordering, got %d", len(q.OrderBy))
				}
				if q.OrderBy[0].Field != "metrics.clicks" {
					t.Errorf("expected metrics.clicks, got %s", q.OrderBy[0].Field)
				}
				if q.OrderBy[0].Direction != Desc {
					t.Errorf("expected DESC, got %s", q.OrderBy[0].Direction)
				}
				return nil
			},
		},
		{
			name:  "with during",
			input: "SELECT campaign.id FROM campaign WHERE segments.date DURING LAST_7_DAYS",
			check: func(q *Query) error {
				if len(q.Where) != 1 {
					t.Errorf("expected 1 condition, got %d", len(q.Where))
				}
				if q.Where[0].Operator != OpDuring {
					t.Errorf("expected DURING, got %s", q.Where[0].Operator)
				}
				if q.Where[0].Value.DateRange != DateRangeLast7Days {
					t.Errorf("expected LAST_7_DAYS, got %s", q.Where[0].Value.DateRange)
				}
				return nil
			},
		},
		{
			name:  "with in clause",
			input: "SELECT campaign.id FROM campaign WHERE campaign.status IN ('ENABLED', 'PAUSED')",
			check: func(q *Query) error {
				if len(q.Where) != 1 {
					t.Errorf("expected 1 condition, got %d", len(q.Where))
				}
				if q.Where[0].Operator != OpIn {
					t.Errorf("expected IN, got %s", q.Where[0].Operator)
				}
				if len(q.Where[0].Value.List) != 2 {
					t.Errorf("expected 2 items, got %d", len(q.Where[0].Value.List))
				}
				return nil
			},
		},
		{
			name:  "multiple where conditions",
			input: "SELECT campaign.id FROM campaign WHERE campaign.status = 'ENABLED' AND metrics.impressions > 0",
			check: func(q *Query) error {
				if len(q.Where) != 2 {
					t.Errorf("expected 2 conditions, got %d", len(q.Where))
				}
				return nil
			},
		},
		{
			name:  "with between",
			input: "SELECT campaign.id FROM campaign WHERE segments.date BETWEEN '2026-01-01' AND '2026-01-31'",
			check: func(q *Query) error {
				if len(q.Where) != 1 {
					t.Errorf("expected 1 condition, got %d", len(q.Where))
				}
				if q.Where[0].Operator != OpBetween {
					t.Errorf("expected BETWEEN, got %s", q.Where[0].Operator)
				}
				if len(q.Where[0].Value.List) != 2 {
					t.Errorf("expected 2 dates, got %d", len(q.Where[0].Value.List))
				}
				return nil
			},
		},
		{
			name:  "numeric comparison",
			input: "SELECT campaign.id FROM campaign WHERE metrics.clicks > 100",
			check: func(q *Query) error {
				if q.Where[0].Operator != OpGt {
					t.Errorf("expected >, got %s", q.Where[0].Operator)
				}
				if q.Where[0].Value.Number != 100 {
					t.Errorf("expected 100, got %f", q.Where[0].Value.Number)
				}
				return nil
			},
		},
		{
			name:  "complex query",
			input: `SELECT campaign.id, campaign.name, metrics.impressions, metrics.clicks
					FROM campaign
					WHERE campaign.status = 'ENABLED'
					  AND segments.date DURING LAST_30_DAYS
					ORDER BY metrics.clicks DESC
					LIMIT 20`,
			check: func(q *Query) error {
				if len(q.Select) != 4 {
					t.Errorf("expected 4 fields, got %d", len(q.Select))
				}
				if len(q.Where) != 2 {
					t.Errorf("expected 2 conditions, got %d", len(q.Where))
				}
				if q.Limit != 20 {
					t.Errorf("expected limit 20, got %d", q.Limit)
				}
				return nil
			},
		},
		// Error cases
		{
			name:    "missing select",
			input:   "FROM campaign",
			wantErr: true,
		},
		{
			name:    "missing from",
			input:   "SELECT campaign.id",
			wantErr: true,
		},
		{
			name:    "empty select",
			input:   "SELECT FROM campaign",
			wantErr: true,
		},
		{
			name:    "invalid limit",
			input:   "SELECT campaign.id FROM campaign LIMIT -1",
			wantErr: true,
		},
		{
			name:    "invalid limit zero",
			input:   "SELECT campaign.id FROM campaign LIMIT 0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := Parse(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if tt.check != nil {
				tt.check(q)
			}
		})
	}
}

func TestLexer(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TokenType
	}{
		{
			name:  "basic tokens",
			input: "SELECT campaign.id FROM campaign",
			expected: []TokenType{
				TokenSelect, TokenIdent, TokenDot, TokenIdent,
				TokenFrom, TokenIdent, TokenEOF,
			},
		},
		{
			name:     "operators",
			input:    "= != > >= < <=",
			expected: []TokenType{TokenEq, TokenNeq, TokenGt, TokenGte, TokenLt, TokenLte, TokenEOF},
		},
		{
			name:     "string literals",
			input:    "'hello' \"world\"",
			expected: []TokenType{TokenString, TokenString, TokenEOF},
		},
		{
			name:     "numbers",
			input:    "123 45.67 -10",
			expected: []TokenType{TokenNumber, TokenNumber, TokenNumber, TokenEOF},
		},
		{
			name:     "date range keywords",
			input:    "DURING LAST_7_DAYS",
			expected: []TokenType{TokenDuring, TokenDateRange, TokenEOF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			tokens, err := lexer.Tokenize()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if len(tokens) != len(tt.expected) {
				t.Errorf("expected %d tokens, got %d", len(tt.expected), len(tokens))
				for i, tok := range tokens {
					t.Logf("token %d: %s %q", i, tok.Type, tok.Value)
				}
				return
			}
			for i, tok := range tokens {
				if tok.Type != tt.expected[i] {
					t.Errorf("token %d: expected %s, got %s", i, tt.expected[i], tok.Type)
				}
			}
		})
	}
}
