// Package main provides the adtap CLI for Google Ads API exploration.
//
// This is a READ-ONLY exploration tool. No mutate operations are supported.
//
// Usage:
//
//	adtap [command] [flags]
//
// Commands:
//
//	search      Execute a GAQL query
//	customers   List accessible customers
//	campaigns   List campaigns for a customer
//	version     Print version information
//
// This tool can be used:
//   - Manually from the command line
//   - Through an LLM integration
//   - Wrapped in an MCP server
package main

import (
	"fmt"
	"os"
)

const (
	version = "0.1.0-alpha"
	name    = "adtap"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	cmd := os.Args[1]

	switch cmd {
	case "version", "-v", "--version":
		printVersion()
	case "help", "-h", "--help":
		printUsage()
	case "search":
		cmdSearch(os.Args[2:])
	case "customers":
		cmdCustomers(os.Args[2:])
	case "campaigns":
		cmdCampaigns(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printVersion() {
	fmt.Printf("%s version %s\n", name, version)
	fmt.Println("Google Ads API v23 - READ-ONLY exploration tool")
}

func printUsage() {
	usage := `adtap - Google Ads API Exploration Tool (READ-ONLY)

Usage:
  adtap <command> [options]

Commands:
  search       Execute a GAQL query against the API
  customers    List accessible customer accounts
  campaigns    List campaigns for a customer
  version      Print version information
  help         Show this help message

Examples:
  adtap customers
  adtap campaigns --customer-id 1234567890
  adtap search --customer-id 1234567890 --query "SELECT campaign.id, campaign.name FROM campaign LIMIT 10"

Environment Variables:
  GOOGLE_ADS_DEVELOPER_TOKEN     Developer token (required)
  GOOGLE_APPLICATION_CREDENTIALS Path to service account JSON
  GOOGLE_PROJECT_ID              GCP project ID

Note: This is a READ-ONLY tool. No mutate operations are supported.
`
	fmt.Print(usage)
}

func cmdSearch(args []string) {
	// TODO: Implement GAQL search
	fmt.Println("search: Not yet implemented")
	fmt.Println("Placeholder for: Execute GAQL query via GoogleAdsService.Search")
}

func cmdCustomers(args []string) {
	// TODO: Implement list accessible customers
	fmt.Println("customers: Not yet implemented")
	fmt.Println("Placeholder for: CustomerService.ListAccessibleCustomers")
}

func cmdCampaigns(args []string) {
	// TODO: Implement list campaigns
	fmt.Println("campaigns: Not yet implemented")
	fmt.Println("Placeholder for: Search campaigns via GAQL")
}
