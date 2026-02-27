# L7 Engineering Review: ADTAP CLI Validation and Contracts

**Date:** 2026-02-27
**Reviewer:** L7 Engineering Lead
**Scope:** CLI validation, contracts, testing strategy
**Project:** adtap (Google Ads API CLI)

---

## Executive Summary

ADTAP is in early development with a solid meta-prompt specification (`docs/meta-prompt.md`) but minimal implementation. The current CLI (`cmd/adtap/main.go`) is a placeholder with stub commands. This review decomposes the work into actionable tasks focused on CLI validation, contracts, and testing.

**Key Findings:**
1. 24 open beads with good thematic coverage but gaps in validation/contracts
2. Meta-prompt is comprehensive but needs formal contract extraction
3. No test infrastructure exists yet
4. Flag compatibility matrix not documented
5. Error handling taxonomy undefined

---

## A. EPIC REVIEW

### A.1 Beads by Theme

#### Theme 1: Core CLI Implementation (2 beads)
| Bead ID | Priority | Title | Status |
|---------|----------|-------|--------|
| adtap-9qj | P1 | Implement adtap CLI per meta-prompt | OPEN |
| adtap-z8j | P2 | Create adtap CLI placeholder | OPEN |

**Gap:** `adtap-z8j` appears already done (main.go exists). Should be closed or refined.

#### Theme 2: Research Tasks (11 beads)
| Bead ID | Priority | Title |
|---------|----------|-------|
| adtap-dao | P1 | Research Performance Max campaigns |
| adtap-jnd | P2 | Research PMax Retail/Lead Gen |
| adtap-0j7 | P2 | Research Bidding strategy types |
| adtap-5bk | P2 | Research Budget tracking |
| adtap-3ow | P2 | Research Ad types |
| adtap-sch | P2 | Research Campaign types |
| adtap-ncm | P2 | Research Asset Management API |
| adtap-8qy | P2 | Research Account Management API |
| adtap-3vd | P2 | Research Go protobuf/HTTP/CLI tools |
| adtap-3kk | P3 | Research Location targeting |
| adtap-izb | P3 | Research Billing API |
| adtap-g15 | P3 | Research Generative Asset Creation |
| adtap-9sg | P4 | Research Campaign drafts |

**Gap:** Research is exploratory. No beads exist for translating research into CLI features.

#### Theme 3: Expert Reviews (4 beads)
| Bead ID | Priority | Title |
|---------|----------|-------|
| adtap-8eq | P2 | Unix CLI conventions/clig.dev compliance |
| adtap-1zb | P2 | Meta-prompt effectiveness |
| adtap-91w | P2 | Go CLI implementation patterns |
| adtap-d56 | P3 | Org-mode documentation structure |

**Gap:** No review bead for contracts or validation logic specifically.

#### Theme 4: Testing & Validation (1 bead)
| Bead ID | Priority | Title |
|---------|----------|-------|
| adtap-2x5 | P2 | Build test harness for incompatible flag combinations |

**Critical Gap:** Only 1 testing bead. No beads for:
- Input validation
- Error code testing
- Contract compliance testing
- GAQL syntax validation

#### Theme 5: Tooling & Documentation (3 beads)
| Bead ID | Priority | Title |
|---------|----------|-------|
| adtap-iaz | P2 | Install Vale and documentation standards |
| adtap-tzj | P3 | Configure Vale contracts for help files |
| adtap-9d1 | P3 | Formal model for GAQL transformations |

#### Theme 6: Integration (1 bead)
| Bead ID | Priority | Title |
|---------|----------|-------|
| adtap-1as | P3 | Design BigQuery integration hooks |

### A.2 Dependency Analysis

```
adtap-9qj (P1: Full CLI implementation)
  |
  +-- depends on --> adtap-8eq (clig.dev compliance review)
  +-- depends on --> adtap-91w (Go patterns review)
  +-- depends on --> adtap-1zb (meta-prompt review)
  +-- depends on --> adtap-2x5 (test harness)
  +-- should create --> contract definitions (MISSING)

adtap-2x5 (P2: Test harness)
  |
  +-- depends on --> flag compatibility matrix (MISSING)
  +-- depends on --> error code taxonomy (MISSING)

adtap-dao (P1: PMax research)
  |
  +-- should inform --> adtap-9qj (CLI feature design)
```

### A.3 Missing Epic-Level Work

1. **Contract Definition Epic** - No beads for defining input/output contracts
2. **Error Handling Epic** - No beads for error taxonomy, exit codes
3. **Integration Testing Epic** - No beads for API mock testing
4. **Configuration Validation Epic** - No beads for config.toml schema validation

---

## B. CLI VALIDATION DECOMPOSITION

### B.1 Flag Compatibility Matrix

Based on meta-prompt analysis (`docs/meta-prompt.md`), here are the expected flag interactions:

#### Global Flags (apply to all commands)

| Flag | Short | Type | Default | Conflicts |
|------|-------|------|---------|-----------|
| `--json` | | bool | false | `--format` (mutual) |
| `--verbose` | `-v` | bool | false | `--quiet` |
| `--quiet` | `-q` | bool | false | `--verbose` |
| `--actor` | | string | "" | none |
| `--config` | | path | auto | none |

#### `query` Command Flags

| Flag | Short | Type | Default | Conflicts | Requires |
|------|-------|------|---------|-----------|----------|
| `--format` | `-f` | enum | table | `--json` | - |
| `--output` | `-o` | path | stdout | - | - |
| `--limit` | | int | 100 | - | - |
| `--dry-run` | | bool | false | - | - |
| `--explain` | | bool | false | - | - |
| `--repl` | | bool | false | ACCOUNT_ID, GAQL args | - |

**Incompatible Combinations:**
```
--repl + (ACCOUNT_ID as positional) = ERROR: repl is interactive
--repl + (--output FILE)            = ERROR: repl writes to stdout interactively
--dry-run + (--output FILE)         = WARNING: no data to write
--json + --format csv               = ERROR: conflicting format specification
--verbose + --quiet                 = ERROR: mutually exclusive
```

#### `export` Command Flags

| Flag | Short | Type | Default | Conflicts | Requires |
|------|-------|------|---------|-----------|----------|
| `--format` | `-f` | enum | jsonl | - | - |
| `--since` | | date | - | - | - |
| `--template` | | string | - | - | - |
| `--dest` | | path | .adtap/exports | - | - |

**Incompatible Combinations:**
```
--since + (no template with date support) = ERROR: template must support incremental
--format parquet + (streaming output)     = ERROR: parquet requires file output
```

#### `auth` Subcommands

| Subcommand | Flags | Notes |
|------------|-------|-------|
| `login` | `--browser` (bool) | Opens browser if TTY |
| `status` | `--json` | Shows auth state |
| `token` | none | Debug only |
| `revoke` | `--confirm` | Requires confirmation |

### B.2 Input Validation Rules

#### Customer ID Validation

```
Format: NNNNNNNNNN (10 digits, no hyphens)
Regex: ^[0-9]{10}$

Accept:
  1234567890

Reject:
  123-456-7890  (contains hyphens)
  12345678901   (11 digits)
  123456789     (9 digits)
  1234567890a   (contains letter)
```

#### GAQL Syntax Validation

```
Required clauses:
  SELECT <field_list>
  FROM <resource>

Optional clauses (order matters):
  WHERE <conditions>
  ORDER BY <field> [ASC|DESC]
  LIMIT <count>
  PARAMETERS <key=value>

Validation rules:
  1. SELECT must contain at least one field
  2. FROM must specify exactly one resource
  3. Fields must be compatible with resource (check via GoogleAdsFieldService)
  4. Metrics require date context (segments.date in SELECT or WHERE)
  5. LIMIT must be positive integer
  6. DURING accepts predefined ranges only
  7. Date format: YYYY-MM-DD
```

#### Date Format Validation

```
Explicit dates: YYYY-MM-DD
  Valid:   2026-02-27
  Invalid: 27-02-2026, 2026/02/27, Feb 27, 2026

DURING keywords (case insensitive):
  TODAY, YESTERDAY, LAST_7_DAYS, LAST_14_DAYS, LAST_30_DAYS,
  THIS_MONTH, LAST_MONTH, THIS_WEEK_SUN_TODAY, THIS_WEEK_MON_TODAY,
  LAST_WEEK_SUN_SAT, LAST_WEEK_MON_SUN, LAST_BUSINESS_WEEK
```

#### Configuration File Validation

```toml
# Required fields
api_version = "v23"   # Must match supported versions
default_account = ""  # Must reference existing [accounts.X] section

# Account section required fields
[accounts.main]
developer_token = ""  # 22-character alphanumeric
customer_id = ""      # 10-digit numeric
```

### B.3 Error Code Taxonomy

Per clig.dev convention and meta-prompt:

| Exit Code | Category | Description | Example |
|-----------|----------|-------------|---------|
| 0 | SUCCESS | Command completed successfully | Query returned results |
| 1 | GENERAL_ERROR | Unspecified error | API returned error |
| 2 | USAGE_ERROR | Invalid command usage | Missing required flag |
| 3 | AUTH_ERROR | Authentication failed | Token expired |
| 4 | API_ERROR | Google Ads API error | GAQL syntax error |
| 5 | CONFIG_ERROR | Configuration invalid | Missing config file |
| 6 | IO_ERROR | File/network error | Cannot write output |
| 7 | VALIDATION_ERROR | Input validation failed | Invalid customer ID |

**Error Message Format:**
```
Error: <category>: <message>
Details: <additional context>
Hint: <suggested fix>

Exit code: N
```

### B.4 Edge Cases

#### Query Edge Cases

| Case | Input | Expected Behavior |
|------|-------|-------------------|
| Empty query | `adtap query 1234567890 ""` | Exit 2: GAQL required |
| Whitespace query | `adtap query 1234567890 "   "` | Exit 2: GAQL required |
| SELECT only | `adtap query 1234567890 "SELECT campaign.id"` | Exit 4: FROM required |
| Unicode in query | `adtap query ... "SELECT ... WHERE campaign.name = ''"` | Pass through to API |
| Very long query | 10KB+ GAQL | Accept (API limit is higher) |
| click_view multi-day | `... segments.date DURING LAST_7_DAYS` | Exit 4: click_view requires single day |

#### Auth Edge Cases

| Case | Scenario | Expected Behavior |
|------|----------|-------------------|
| No credentials | Fresh install | Exit 5: Run `adtap auth login` |
| Expired token | Refresh token valid | Auto-refresh, continue |
| Expired refresh | Must re-auth | Exit 3: Run `adtap auth login` |
| Revoked access | Token revoked externally | Exit 3: Access revoked |
| No network | Offline | Exit 6: Network error |

#### Output Edge Cases

| Case | Scenario | Expected Behavior |
|------|----------|-------------------|
| Write to non-existent dir | `--output /no/such/dir/file.json` | Exit 6: Directory not found |
| Write to read-only | `--output /etc/file.json` | Exit 6: Permission denied |
| Output file exists | `--output existing.json` | Overwrite (no prompt if not TTY) |
| Disk full | Writing large export | Exit 6: Disk full |
| Stdout piped | `adtap query ... \| jq` | Detect non-TTY, no color |

---

## C. CONTRACT DEFINITIONS

### C.1 Input Contracts

#### GAQL Query Contract

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "adtap/gaql-query",
  "title": "GAQL Query",
  "type": "object",
  "required": ["select", "from"],
  "properties": {
    "select": {
      "type": "array",
      "minItems": 1,
      "items": {
        "type": "string",
        "pattern": "^[a-z_]+\\.[a-z_]+$"
      }
    },
    "from": {
      "type": "string",
      "enum": ["campaign", "ad_group", "ad_group_ad", "customer", ...]
    },
    "where": {
      "type": "array",
      "items": { "$ref": "#/$defs/condition" }
    },
    "orderBy": {
      "type": "object",
      "properties": {
        "field": { "type": "string" },
        "direction": { "enum": ["ASC", "DESC"] }
      }
    },
    "limit": {
      "type": "integer",
      "minimum": 1,
      "maximum": 10000
    }
  }
}
```

#### Customer ID Contract

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "adtap/customer-id",
  "title": "Google Ads Customer ID",
  "type": "string",
  "pattern": "^[0-9]{10}$",
  "description": "10-digit numeric customer ID without hyphens"
}
```

#### Date Range Contract

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "adtap/date-range",
  "oneOf": [
    {
      "type": "string",
      "enum": ["TODAY", "YESTERDAY", "LAST_7_DAYS", "LAST_14_DAYS",
               "LAST_30_DAYS", "THIS_MONTH", "LAST_MONTH"]
    },
    {
      "type": "object",
      "required": ["start", "end"],
      "properties": {
        "start": { "type": "string", "format": "date" },
        "end": { "type": "string", "format": "date" }
      }
    }
  ]
}
```

### C.2 Output Contracts

#### JSON Output Contract (--json flag)

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "adtap/json-output",
  "type": "object",
  "required": ["data", "metadata"],
  "properties": {
    "data": {
      "type": "array",
      "items": { "type": "object" }
    },
    "metadata": {
      "type": "object",
      "required": ["query", "customer_id", "timestamp", "row_count"],
      "properties": {
        "query": { "type": "string" },
        "customer_id": { "type": "string" },
        "timestamp": { "type": "string", "format": "date-time" },
        "row_count": { "type": "integer" },
        "execution_time_ms": { "type": "integer" }
      }
    }
  }
}
```

#### Error Output Contract

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "adtap/error-output",
  "type": "object",
  "required": ["error", "exit_code"],
  "properties": {
    "error": {
      "type": "object",
      "required": ["category", "message"],
      "properties": {
        "category": {
          "enum": ["GENERAL_ERROR", "USAGE_ERROR", "AUTH_ERROR",
                   "API_ERROR", "CONFIG_ERROR", "IO_ERROR", "VALIDATION_ERROR"]
        },
        "message": { "type": "string" },
        "details": { "type": "string" },
        "hint": { "type": "string" }
      }
    },
    "exit_code": {
      "type": "integer",
      "minimum": 1,
      "maximum": 7
    }
  }
}
```

#### Exit Code Contract

| Code | JSON category | Human message prefix |
|------|---------------|----------------------|
| 0 | (no error) | (no message) |
| 1 | GENERAL_ERROR | "Error: " |
| 2 | USAGE_ERROR | "Usage error: " |
| 3 | AUTH_ERROR | "Authentication error: " |
| 4 | API_ERROR | "API error: " |
| 5 | CONFIG_ERROR | "Configuration error: " |
| 6 | IO_ERROR | "I/O error: " |
| 7 | VALIDATION_ERROR | "Validation error: " |

### C.3 Configuration Contract

#### .adtap/config.toml Schema

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "adtap/config-toml",
  "type": "object",
  "required": ["api_version", "default_account", "accounts"],
  "properties": {
    "api_version": {
      "type": "string",
      "pattern": "^v[0-9]+$",
      "description": "API version (e.g., v23)"
    },
    "default_account": {
      "type": "string",
      "description": "Key in accounts table"
    },
    "accounts": {
      "type": "object",
      "additionalProperties": {
        "$ref": "#/$defs/account"
      }
    },
    "defaults": {
      "type": "object",
      "properties": {
        "format": { "enum": ["table", "json", "jsonl", "csv", "parquet"] },
        "limit": { "type": "integer", "minimum": 1 },
        "output_dir": { "type": "string" }
      }
    }
  },
  "$defs": {
    "account": {
      "type": "object",
      "required": ["developer_token", "customer_id"],
      "properties": {
        "developer_token": { "type": "string", "pattern": "^[A-Za-z0-9_-]{22}$" },
        "customer_id": { "type": "string", "pattern": "^[0-9]{10}$" },
        "login_customer_id": { "type": "string", "pattern": "^[0-9]{10}$" },
        "oauth": {
          "type": "object",
          "required": ["client_id", "client_secret"],
          "properties": {
            "client_id": { "type": "string" },
            "client_secret": { "type": "string" }
          }
        }
      }
    }
  }
}
```

### C.4 API Response Contracts

#### SearchStream Response (expected from Google)

```json
{
  "type": "array",
  "description": "Streaming NDJSON batches",
  "items": {
    "type": "object",
    "properties": {
      "results": {
        "type": "array",
        "items": {
          "type": "object",
          "description": "Fields vary by FROM resource"
        }
      },
      "fieldMask": { "type": "string" },
      "requestId": { "type": "string" }
    }
  }
}
```

#### GoogleAdsFieldService Response

```json
{
  "type": "object",
  "properties": {
    "name": { "type": "string" },
    "category": { "enum": ["RESOURCE", "ATTRIBUTE", "METRIC", "SEGMENT"] },
    "dataType": { "type": "string" },
    "selectable": { "type": "boolean" },
    "filterable": { "type": "boolean" },
    "sortable": { "type": "boolean" },
    "selectableWith": { "type": "array", "items": { "type": "string" } },
    "attributeResources": { "type": "array", "items": { "type": "string" } },
    "metrics": { "type": "array", "items": { "type": "string" } },
    "segments": { "type": "array", "items": { "type": "string" } }
  }
}
```

---

## D. RECOMMENDED NEW BEADS

### D.1 P1 (Critical Path)

| ID | Title | Description | Blocks |
|----|-------|-------------|--------|
| NEW-01 | Define CLI flag compatibility matrix | Document all flag conflicts per command. Input for test harness. | adtap-2x5 |
| NEW-02 | Define exit code taxonomy | Formal error categories and exit codes per clig.dev. | adtap-9qj |
| NEW-03 | Create GAQL parser/validator module | Parse and validate GAQL syntax before API call. | adtap-9qj |

### D.2 P2 (High Priority)

| ID | Title | Description | Depends On |
|----|-------|-------------|------------|
| NEW-04 | Define JSON output schema contract | JSON Schema for `--json` output across commands | - |
| NEW-05 | Define config.toml schema contract | JSON Schema for configuration validation | - |
| NEW-06 | Create input validation module | Customer ID, date format, path validation | - |
| NEW-07 | Build API mock infrastructure | httptest mocks for GoogleAdsService responses | - |
| NEW-08 | Create golden file test fixtures | Expected outputs for regression testing | NEW-07 |
| NEW-09 | Review: Error message consistency | Audit all error messages for clig.dev compliance | NEW-02 |
| NEW-10 | Define GAQL template schema | Schema for .gaql template files | - |

### D.3 P3 (Standard Priority)

| ID | Title | Description | Depends On |
|----|-------|-------------|------------|
| NEW-11 | Add pre-commit GAQL syntax check | Validate .gaql templates on commit | NEW-03 |
| NEW-12 | Create CLI completion tests | Test bash/zsh/fish completion scripts | adtap-9qj |
| NEW-13 | Document micros conversion contract | Standard for displaying monetary values | - |
| NEW-14 | Define TTY detection contract | When to use color, interactive prompts | - |
| NEW-15 | Build config migration tool | Migrate config between API versions | NEW-05 |

### D.4 P4 (Nice to Have)

| ID | Title | Description |
|----|-------|-------------|
| NEW-16 | Add GAQL field compatibility cache | Cache selectableWith locally for faster validation |
| NEW-17 | Create query cost estimator | Estimate API quota usage before execution |
| NEW-18 | Add query history/audit log | Log queries for debugging and replay |

---

## E. TESTING STRATEGY

### E.1 Test Categories

```
tests/
├── unit/                    # Pure function tests
│   ├── gaql_parser_test.go  # GAQL parsing/validation
│   ├── validation_test.go   # Input validation
│   ├── config_test.go       # Config parsing
│   └── output_test.go       # Output formatting
├── integration/             # Tests requiring mocks
│   ├── api_mock_test.go     # Mocked API calls
│   ├── auth_flow_test.go    # OAuth mock flow
│   └── export_test.go       # Export workflows
├── e2e/                     # End-to-end (gated by ADTAP_INTEGRATION=1)
│   └── real_api_test.go     # Real API calls (test account)
├── golden/                  # Golden file tests
│   ├── testdata/
│   │   ├── query_output.golden
│   │   └── error_output.golden
│   └── golden_test.go
└── fuzz/                    # Fuzz testing
    └── gaql_fuzz_test.go    # Fuzz GAQL parser
```

### E.2 Flag Combination Test Matrix

```go
// tests/unit/flags_test.go
var flagConflictTests = []struct {
    name     string
    flags    []string
    wantErr  bool
    exitCode int
}{
    // Global flag conflicts
    {"verbose and quiet", []string{"-v", "-q"}, true, 2},
    {"json and format", []string{"--json", "--format", "csv"}, true, 2},

    // Query command conflicts
    {"repl with output file", []string{"query", "--repl", "-o", "file.json"}, true, 2},
    {"repl with positional GAQL", []string{"query", "123", "SELECT...", "--repl"}, true, 2},
    {"dry-run with output", []string{"query", "--dry-run", "-o", "file.json"}, false, 0}, // warn only

    // Valid combinations
    {"verbose with json", []string{"-v", "--json"}, false, 0},
    {"query with all options", []string{"query", "123", "SELECT...", "-f", "jsonl", "-o", "out.jsonl"}, false, 0},
}
```

### E.3 Golden File Approach

```go
// tests/golden/golden_test.go
func TestQueryOutput(t *testing.T) {
    golden.Run(t, "testdata/query_output", func(t *testing.T, input []byte) []byte {
        // Run command with input, capture output
        return runCommand("query", "1234567890", string(input))
    })
}
```

### E.4 Coverage Targets

| Category | Target | Rationale |
|----------|--------|-----------|
| Unit tests | 80% | Core logic must be tested |
| Flag parsing | 100% | All combinations documented |
| Error paths | 90% | Error handling is critical for CLI |
| Happy paths | 100% | All documented use cases |

---

## F. IMPLEMENTATION PRIORITIES

### Phase 1: Contracts and Validation (P1-P2)
1. NEW-02: Exit code taxonomy
2. NEW-01: Flag compatibility matrix
3. NEW-03: GAQL parser/validator
4. NEW-04: JSON output schema
5. NEW-05: Config schema

### Phase 2: Test Infrastructure (P2)
1. adtap-2x5: Test harness for flags
2. NEW-07: API mock infrastructure
3. NEW-08: Golden file fixtures

### Phase 3: Implementation (P1)
1. adtap-9qj: Full CLI implementation (depends on Phase 1-2)

### Phase 4: Expert Reviews (P2)
1. adtap-8eq: clig.dev compliance
2. adtap-91w: Go patterns
3. NEW-09: Error message consistency

---

## G. APPENDIX: Bead Status Recommendations

### Close (already done)
- `adtap-z8j` - CLI placeholder exists in main.go

### Update (needs refinement)
- `adtap-2x5` - Add dependency on NEW-01 (flag matrix)
- `adtap-9qj` - Add dependencies on contract beads

### Blocked (needs prerequisites)
- `adtap-9qj` - Blocked by: NEW-01, NEW-02, NEW-03

---

*End of L7 Engineering Review*
