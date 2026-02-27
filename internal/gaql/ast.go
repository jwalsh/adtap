// Package gaql provides parsing and validation for Google Ads Query Language.
package gaql

import (
	"fmt"
	"strings"
)

// Query represents a parsed GAQL query.
type Query struct {
	Select     []Field
	From       string
	Where      []Condition
	OrderBy    []Ordering
	Limit      int
	Parameters map[string]string
}

// Field represents a field reference (e.g., campaign.id, metrics.clicks).
type Field struct {
	Name string
}

// Condition represents a WHERE clause condition.
type Condition struct {
	Field    string
	Operator Operator
	Value    Value
}

// Ordering represents an ORDER BY clause item.
type Ordering struct {
	Field     string
	Direction Direction
}

// Direction represents sort direction.
type Direction int

const (
	Asc Direction = iota
	Desc
)

func (d Direction) String() string {
	switch d {
	case Desc:
		return "DESC"
	default:
		return "ASC"
	}
}

// Operator represents a comparison operator.
type Operator int

const (
	OpEq Operator = iota
	OpNeq
	OpGt
	OpGte
	OpLt
	OpLte
	OpIn
	OpNotIn
	OpLike
	OpNotLike
	OpContainsAny
	OpContainsAll
	OpContainsNone
	OpIsNull
	OpIsNotNull
	OpDuring
	OpBetween
	OpRegexpMatch
	OpNotRegexpMatch
)

func (o Operator) String() string {
	switch o {
	case OpEq:
		return "="
	case OpNeq:
		return "!="
	case OpGt:
		return ">"
	case OpGte:
		return ">="
	case OpLt:
		return "<"
	case OpLte:
		return "<="
	case OpIn:
		return "IN"
	case OpNotIn:
		return "NOT IN"
	case OpLike:
		return "LIKE"
	case OpNotLike:
		return "NOT LIKE"
	case OpContainsAny:
		return "CONTAINS ANY"
	case OpContainsAll:
		return "CONTAINS ALL"
	case OpContainsNone:
		return "CONTAINS NONE"
	case OpIsNull:
		return "IS NULL"
	case OpIsNotNull:
		return "IS NOT NULL"
	case OpDuring:
		return "DURING"
	case OpBetween:
		return "BETWEEN"
	case OpRegexpMatch:
		return "REGEXP_MATCH"
	case OpNotRegexpMatch:
		return "NOT REGEXP_MATCH"
	default:
		return "UNKNOWN"
	}
}

// Value represents a value in a condition.
type Value struct {
	Type      ValueType
	Str       string // String value (renamed from String to avoid method conflict)
	Number    float64
	List      []string
	DateRange DateRange
}

// ValueType represents the type of a value.
type ValueType int

const (
	ValueString ValueType = iota
	ValueNumber
	ValueList
	ValueDateRange
	ValueNull
)

// DateRange represents a DURING clause date range.
type DateRange int

const (
	DateRangeToday DateRange = iota
	DateRangeYesterday
	DateRangeLast7Days
	DateRangeLast14Days
	DateRangeLast30Days
	DateRangeThisMonth
	DateRangeLastMonth
	DateRangeThisWeekSunToday
	DateRangeThisWeekMonToday
	DateRangeLastWeekSunSat
	DateRangeLastWeekMonSun
	DateRangeLastBusinessWeek
	DateRangeCustom // For BETWEEN date ranges
)

// DateRangeKeywords maps string keywords to DateRange values.
var DateRangeKeywords = map[string]DateRange{
	"TODAY":               DateRangeToday,
	"YESTERDAY":           DateRangeYesterday,
	"LAST_7_DAYS":         DateRangeLast7Days,
	"LAST_14_DAYS":        DateRangeLast14Days,
	"LAST_30_DAYS":        DateRangeLast30Days,
	"THIS_MONTH":          DateRangeThisMonth,
	"LAST_MONTH":          DateRangeLastMonth,
	"THIS_WEEK_SUN_TODAY": DateRangeThisWeekSunToday,
	"THIS_WEEK_MON_TODAY": DateRangeThisWeekMonToday,
	"LAST_WEEK_SUN_SAT":   DateRangeLastWeekSunSat,
	"LAST_WEEK_MON_SUN":   DateRangeLastWeekMonSun,
	"LAST_BUSINESS_WEEK":  DateRangeLastBusinessWeek,
}

func (d DateRange) String() string {
	for k, v := range DateRangeKeywords {
		if v == d {
			return k
		}
	}
	return "CUSTOM"
}

// String returns the GAQL query as a string.
func (q *Query) String() string {
	var sb strings.Builder

	// SELECT
	sb.WriteString("SELECT ")
	for i, f := range q.Select {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(f.Name)
	}

	// FROM
	sb.WriteString(" FROM ")
	sb.WriteString(q.From)

	// WHERE
	if len(q.Where) > 0 {
		sb.WriteString(" WHERE ")
		for i, c := range q.Where {
			if i > 0 {
				sb.WriteString(" AND ")
			}
			sb.WriteString(c.Field)
			sb.WriteString(" ")
			sb.WriteString(c.Operator.String())
			sb.WriteString(" ")
			sb.WriteString(c.Value.String())
		}
	}

	// ORDER BY
	if len(q.OrderBy) > 0 {
		sb.WriteString(" ORDER BY ")
		for i, o := range q.OrderBy {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(o.Field)
			if o.Direction == Desc {
				sb.WriteString(" DESC")
			}
		}
	}

	// LIMIT
	if q.Limit > 0 {
		sb.WriteString(fmt.Sprintf(" LIMIT %d", q.Limit))
	}

	// PARAMETERS
	if len(q.Parameters) > 0 {
		sb.WriteString(" PARAMETERS ")
		first := true
		for k, v := range q.Parameters {
			if !first {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%s = %s", k, v))
			first = false
		}
	}

	return sb.String()
}

// String returns the value as a string representation.
func (v Value) String() string {
	switch v.Type {
	case ValueString:
		return fmt.Sprintf("'%s'", v.Str)
	case ValueNumber:
		return fmt.Sprintf("%v", v.Number)
	case ValueList:
		return fmt.Sprintf("(%s)", strings.Join(v.List, ", "))
	case ValueDateRange:
		return v.DateRange.String()
	case ValueNull:
		return "NULL"
	default:
		return ""
	}
}
