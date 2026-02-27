package gaql

import (
	"regexp"
	"strings"
)

// KnownResources lists the common Google Ads API resources.
// This is not exhaustive; the API has many more resources.
var KnownResources = map[string]bool{
	"campaign":                       true,
	"ad_group":                       true,
	"ad_group_ad":                    true,
	"ad_group_criterion":             true,
	"asset":                          true,
	"campaign_asset":                 true,
	"campaign_budget":                true,
	"campaign_criterion":             true,
	"customer":                       true,
	"customer_client":                true,
	"change_event":                   true,
	"change_status":                  true,
	"click_view":                     true,
	"conversion_action":              true,
	"geo_target_constant":            true,
	"keyword_view":                   true,
	"label":                          true,
	"location_view":                  true,
	"media_file":                     true,
	"mobile_app_category_constant":   true,
	"mobile_device_constant":         true,
	"performance_max_placement_view": true,
	"product_bidding_category_constant": true,
	"search_term_view":               true,
	"shopping_performance_view":      true,
	"topic_constant":                 true,
	"user_list":                      true,
}

// SingleDayResources are resources that require single-day date queries.
var SingleDayResources = map[string]bool{
	"click_view": true,
}

// FieldCategories maps field prefixes to their categories.
var FieldCategories = map[string]string{
	"metrics":  "METRIC",
	"segments": "SEGMENT",
}

// datePattern matches YYYY-MM-DD format.
var datePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// Validator performs semantic validation on parsed GAQL queries.
type Validator struct {
	// AllowUnknownResources permits resources not in KnownResources.
	// Useful for newer API resources not yet in the list.
	AllowUnknownResources bool

	// RequireMetricDateContext enforces that metrics require date segments.
	RequireMetricDateContext bool
}

// NewValidator creates a new validator with default settings.
func NewValidator() *Validator {
	return &Validator{
		AllowUnknownResources:    true, // Default permissive for forward compat
		RequireMetricDateContext: true,
	}
}

// Validate performs semantic validation on a parsed query.
func (v *Validator) Validate(q *Query) error {
	if err := v.validateSelect(q); err != nil {
		return err
	}
	if err := v.validateFrom(q); err != nil {
		return err
	}
	if err := v.validateWhere(q); err != nil {
		return err
	}
	if err := v.validateLimit(q); err != nil {
		return err
	}
	if err := v.validateSingleDayResource(q); err != nil {
		return err
	}
	if err := v.validateMetricDateContext(q); err != nil {
		return err
	}
	return nil
}

func (v *Validator) validateSelect(q *Query) error {
	if len(q.Select) == 0 {
		return &ValidationError{Message: "SELECT must contain at least one field"}
	}

	for _, f := range q.Select {
		if err := v.validateFieldName(f.Name); err != nil {
			return err
		}
	}

	return nil
}

func (v *Validator) validateFrom(q *Query) error {
	if q.From == "" {
		return &ValidationError{Message: "FROM clause is required"}
	}

	if !v.AllowUnknownResources {
		if _, ok := KnownResources[q.From]; !ok {
			return &ValidationError{
				Message: "unknown resource: " + q.From,
				Field:   "FROM",
			}
		}
	}

	return nil
}

func (v *Validator) validateWhere(q *Query) error {
	for _, cond := range q.Where {
		if err := v.validateFieldName(cond.Field); err != nil {
			return err
		}

		// Validate DURING date ranges
		if cond.Operator == OpDuring {
			if cond.Value.Type != ValueDateRange {
				return &ValidationError{
					Message: "DURING requires a date range keyword",
					Field:   cond.Field,
				}
			}
		}

		// Validate BETWEEN dates
		if cond.Operator == OpBetween {
			if cond.Value.Type != ValueList || len(cond.Value.List) != 2 {
				return &ValidationError{
					Message: "BETWEEN requires two values",
					Field:   cond.Field,
				}
			}
			for _, d := range cond.Value.List {
				if !datePattern.MatchString(d) && !isDateRangeKeyword(d) {
					return &ValidationError{
						Message: "invalid date format (expected YYYY-MM-DD): " + d,
						Field:   cond.Field,
					}
				}
			}
		}
	}

	return nil
}

func (v *Validator) validateLimit(q *Query) error {
	if q.Limit < 0 {
		return &ValidationError{Message: "LIMIT must be non-negative"}
	}
	return nil
}

func (v *Validator) validateSingleDayResource(q *Query) error {
	if !SingleDayResources[q.From] {
		return nil
	}

	// click_view requires single-day queries
	for _, cond := range q.Where {
		if cond.Field == "segments.date" {
			if cond.Operator == OpDuring {
				dr := cond.Value.DateRange
				if dr == DateRangeToday || dr == DateRangeYesterday {
					return nil
				}
				return &ValidationError{
					Message: "click_view requires single-day date range (TODAY or YESTERDAY)",
					Field:   "segments.date",
				}
			}
			if cond.Operator == OpEq {
				return nil // Single day via equality
			}
			if cond.Operator == OpBetween {
				// Check if start == end
				if len(cond.Value.List) == 2 && cond.Value.List[0] == cond.Value.List[1] {
					return nil
				}
				return &ValidationError{
					Message: "click_view requires single-day date range",
					Field:   "segments.date",
				}
			}
		}
	}

	return &ValidationError{
		Message: "click_view requires segments.date in WHERE clause with single-day range",
		Field:   "FROM",
	}
}

func (v *Validator) validateMetricDateContext(q *Query) error {
	if !v.RequireMetricDateContext {
		return nil
	}

	hasMetrics := false
	for _, f := range q.Select {
		if strings.HasPrefix(f.Name, "metrics.") {
			hasMetrics = true
			break
		}
	}

	if !hasMetrics {
		return nil
	}

	// Check for date context in SELECT or WHERE
	hasDateContext := false

	for _, f := range q.Select {
		if f.Name == "segments.date" {
			hasDateContext = true
			break
		}
	}

	if !hasDateContext {
		for _, cond := range q.Where {
			if cond.Field == "segments.date" {
				hasDateContext = true
				break
			}
		}
	}

	if !hasDateContext {
		return &ValidationError{
			Message: "metrics require date context (segments.date in SELECT or WHERE)",
		}
	}

	return nil
}

func (v *Validator) validateFieldName(name string) error {
	if name == "" {
		return &ValidationError{Message: "field name cannot be empty"}
	}

	// Field names should contain at least one dot for qualified names
	// e.g., campaign.id, metrics.clicks
	// Single-part names are also valid (e.g., for resources)

	return nil
}

func isDateRangeKeyword(s string) bool {
	_, ok := DateRangeKeywords[strings.ToUpper(s)]
	return ok
}

// ValidateQuery parses and validates a GAQL query string.
func ValidateQuery(input string) (*Query, error) {
	q, err := Parse(input)
	if err != nil {
		return nil, err
	}

	v := NewValidator()
	if err := v.Validate(q); err != nil {
		return nil, err
	}

	return q, nil
}
