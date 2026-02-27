// Package gaql provides parsing and validation for Google Ads Query Language (GAQL).
//
// GAQL is a SQL-like query language used to retrieve data from the Google Ads API.
// This package parses GAQL queries into an AST and validates them before API calls.
//
// # Basic Usage
//
//	q, err := gaql.Parse("SELECT campaign.id FROM campaign WHERE campaign.status = 'ENABLED'")
//	if err != nil {
//		log.Fatal(err)
//	}
//	// Use q.Select, q.From, q.Where, etc.
//
// # Validation
//
// The ValidateQuery function parses and validates a query:
//
//	q, err := gaql.ValidateQuery("SELECT campaign.id, metrics.clicks FROM campaign WHERE segments.date DURING LAST_7_DAYS")
//	if err != nil {
//		log.Fatal(err)
//	}
//
// Validation checks include:
//   - Required SELECT and FROM clauses
//   - Valid operators and date range keywords
//   - Metrics require date context (segments.date)
//   - Single-day resources (click_view) require single-day date ranges
//
// # Custom Validation
//
// For more control, use the Validator directly:
//
//	q, err := gaql.Parse(input)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	v := gaql.NewValidator()
//	v.AllowUnknownResources = false  // Strict mode
//	v.RequireMetricDateContext = true
//
//	if err := v.Validate(q); err != nil {
//		log.Fatal(err)
//	}
//
// # Query Structure
//
// A GAQL query has the following structure:
//
//	SELECT field1, field2, ...
//	FROM resource
//	WHERE condition1 AND condition2 ...
//	ORDER BY field [ASC|DESC]
//	LIMIT count
//	PARAMETERS key=value, ...
//
// Only SELECT and FROM are required. All other clauses are optional.
//
// # Supported Operators
//
// Comparison: =, !=, >, >=, <, <=
// Set: IN, NOT IN
// Pattern: LIKE, NOT LIKE, REGEXP_MATCH, NOT REGEXP_MATCH
// Contains: CONTAINS ANY, CONTAINS ALL, CONTAINS NONE
// Null: IS NULL, IS NOT NULL
// Date: DURING, BETWEEN
//
// # Date Ranges
//
// The DURING operator accepts predefined date ranges:
//
//	TODAY, YESTERDAY
//	LAST_7_DAYS, LAST_14_DAYS, LAST_30_DAYS
//	THIS_MONTH, LAST_MONTH
//	THIS_WEEK_SUN_TODAY, THIS_WEEK_MON_TODAY
//	LAST_WEEK_SUN_SAT, LAST_WEEK_MON_SUN
//	LAST_BUSINESS_WEEK
//
// For custom ranges, use BETWEEN with dates in YYYY-MM-DD format:
//
//	WHERE segments.date BETWEEN '2026-01-01' AND '2026-01-31'
package gaql
