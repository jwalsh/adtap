package gaql

import (
	"strings"
	"testing"
)

func TestValidateQuery(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		errMsg  string
	}{
		{
			name:  "valid simple query",
			input: "SELECT campaign.id, campaign.name FROM campaign",
		},
		{
			name:  "valid query with metrics and date",
			input: "SELECT campaign.id, metrics.clicks FROM campaign WHERE segments.date DURING LAST_7_DAYS",
		},
		{
			name:  "valid query with metrics and date in select",
			input: "SELECT campaign.id, segments.date, metrics.clicks FROM campaign",
		},
		{
			name:    "metrics without date context",
			input:   "SELECT campaign.id, metrics.clicks FROM campaign",
			wantErr: true,
			errMsg:  "metrics require date context",
		},
		{
			name:    "click_view with multi-day range",
			input:   "SELECT click_view.gclid FROM click_view WHERE segments.date DURING LAST_7_DAYS",
			wantErr: true,
			errMsg:  "click_view requires single-day date range",
		},
		{
			name:  "click_view with TODAY",
			input: "SELECT click_view.gclid FROM click_view WHERE segments.date DURING TODAY",
		},
		{
			name:  "click_view with YESTERDAY",
			input: "SELECT click_view.gclid FROM click_view WHERE segments.date DURING YESTERDAY",
		},
		{
			name:  "click_view with date equality",
			input: "SELECT click_view.gclid FROM click_view WHERE segments.date = '2026-02-27'",
		},
		{
			name:  "valid between dates",
			input: "SELECT campaign.id FROM campaign WHERE segments.date BETWEEN '2026-01-01' AND '2026-01-31'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateQuery(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateSelectFields(t *testing.T) {
	tests := []struct {
		name    string
		query   *Query
		wantErr bool
	}{
		{
			name: "valid fields",
			query: &Query{
				Select: []Field{{Name: "campaign.id"}, {Name: "campaign.name"}},
				From:   "campaign",
			},
		},
		{
			name: "empty select",
			query: &Query{
				Select: []Field{},
				From:   "campaign",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.RequireMetricDateContext = false
			err := v.Validate(tt.query)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateResource(t *testing.T) {
	tests := []struct {
		name      string
		resource  string
		allowUnkn bool
		wantErr   bool
	}{
		{
			name:     "known resource",
			resource: "campaign",
		},
		{
			name:     "known resource ad_group",
			resource: "ad_group",
		},
		{
			name:      "unknown resource with allowUnknown true",
			resource:  "new_resource_v99",
			allowUnkn: true,
		},
		{
			name:      "unknown resource with allowUnknown false",
			resource:  "new_resource_v99",
			allowUnkn: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.AllowUnknownResources = tt.allowUnkn
			v.RequireMetricDateContext = false
			q := &Query{
				Select: []Field{{Name: tt.resource + ".id"}},
				From:   tt.resource,
			}
			err := v.Validate(q)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestParseAndValidate(t *testing.T) {
	// Integration test for common GAQL patterns from the documentation
	queries := []string{
		// Campaign overview
		`SELECT
		  campaign.id,
		  campaign.name,
		  campaign.status,
		  campaign.advertising_channel_type,
		  campaign_budget.amount_micros,
		  metrics.impressions,
		  metrics.clicks,
		  metrics.conversions
		FROM campaign
		WHERE segments.date DURING LAST_30_DAYS
		  AND campaign.status != 'REMOVED'
		ORDER BY metrics.impressions DESC`,

		// Ad group performance
		`SELECT
		  ad_group.id,
		  ad_group.name,
		  ad_group.status,
		  campaign.name,
		  metrics.impressions,
		  metrics.clicks,
		  metrics.ctr
		FROM ad_group
		WHERE segments.date DURING LAST_30_DAYS
		ORDER BY metrics.clicks DESC
		LIMIT 20`,

		// Filter by status
		`SELECT campaign.id, campaign.name
		FROM campaign
		WHERE campaign.status = 'ENABLED'`,

		// In clause
		`SELECT campaign.id, campaign.name
		FROM campaign
		WHERE campaign.status IN ('ENABLED', 'PAUSED')`,

		// Segmentation
		`SELECT
		  campaign.name,
		  segments.date,
		  segments.device,
		  metrics.clicks
		FROM campaign
		WHERE segments.date DURING LAST_7_DAYS`,
	}

	for i, q := range queries {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			_, err := ValidateQuery(q)
			if err != nil {
				t.Errorf("query %d failed validation: %v", i, err)
				t.Logf("query: %s", q)
			}
		})
	}
}
