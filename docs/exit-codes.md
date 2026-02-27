# adtap Exit Code Taxonomy

This document defines the exit code conventions for `adtap`, per [clig.dev](https://clig.dev/) guidelines.

## Exit Code Table

| Code | Category | Constant | Description |
|------|----------|----------|-------------|
| 0 | SUCCESS | `ExitSuccess` | Command completed successfully |
| 1 | GENERAL_ERROR | `ExitGeneralError` | Unspecified error |
| 2 | USAGE_ERROR | `ExitUsageError` | Invalid command usage or arguments |
| 3 | AUTH_ERROR | `ExitAuthError` | Authentication or authorization failed |
| 4 | API_ERROR | `ExitAPIError` | Google Ads API returned an error |
| 5 | CONFIG_ERROR | `ExitConfigError` | Configuration invalid or missing |
| 6 | IO_ERROR | `ExitIOError` | File or network I/O error |
| 7 | VALIDATION_ERROR | `ExitValidationError` | Input validation failed |

## Exit Code Details

### 0 - SUCCESS

Command executed successfully. Data output (if any) written to stdout.

**Examples:**
- Query returned results
- Configuration displayed
- Authentication succeeded
- Export completed

### 1 - GENERAL_ERROR

Catch-all for errors that don't fit other categories. Avoid using this code when a more specific category applies.

**Examples:**
- Unexpected internal error
- Unhandled exception
- Unknown error from external system

**Error message format:**
```
Error: <message>
```

### 2 - USAGE_ERROR

Invalid command-line usage: missing arguments, unknown flags, or incompatible flag combinations.

**Examples:**
- Missing required positional argument: `adtap query` (no account ID)
- Unknown flag: `adtap query --badflg`
- Incompatible flags: `adtap query --verbose --quiet`
- Conflicting format flags: `adtap query --json --format csv`
- REPL with output file: `adtap query --repl -o file.json`

**Error message format:**
```
Usage error: <message>

Run 'adtap <command> --help' for usage.
```

### 3 - AUTH_ERROR

Authentication or authorization failed. Credentials missing, expired, or revoked.

**Examples:**
- No credentials configured
- OAuth token expired and refresh failed
- Access token revoked
- Developer token invalid
- Insufficient API access level
- Customer ID not accessible

**Error message format:**
```
Authentication error: <message>

Hint: Run 'adtap auth login' to authenticate.
```

### 4 - API_ERROR

Google Ads API returned an error. Includes GAQL syntax errors, quota errors, and server errors.

**Examples:**
- GAQL syntax error: `SELECT * FROM campaign` (no * in GAQL)
- Invalid field: `SELECT campaign.nonexistent FROM campaign`
- Incompatible fields: field not selectable with resource
- Rate limit exceeded
- API server error (5xx)
- Request too large

**Error message format:**
```
API error: <message>

GAQL: <query if applicable>
Request ID: <request_id>
```

### 5 - CONFIG_ERROR

Configuration file missing, malformed, or contains invalid values.

**Examples:**
- Config file not found
- TOML parse error
- Missing required field: `developer_token`
- Invalid API version: `v999`
- Referenced account not defined
- Credentials file corrupt

**Error message format:**
```
Configuration error: <message>

Config file: <path>
Hint: Run 'adtap config show' to inspect configuration.
```

### 6 - IO_ERROR

File or network I/O operation failed.

**Examples:**
- Cannot write to output file (permission denied)
- Output directory does not exist
- Disk full
- Network unreachable
- DNS resolution failed
- Connection timeout
- TLS handshake failed

**Error message format:**
```
I/O error: <message>

Path: <file path if applicable>
```

### 7 - VALIDATION_ERROR

Input validation failed before API call. Distinct from API_ERROR (exit code 4) which occurs after the API rejects the request.

**Examples:**
- Invalid customer ID format: `123-456-7890` (should be `1234567890`)
- Invalid date format: `27-02-2026` (should be `2026-02-27`)
- Invalid DURING keyword
- Empty GAQL query
- Invalid output format: `--format xml`

**Error message format:**
```
Validation error: <message>

Expected: <correct format>
Got: <actual value>
```

## Error Message Guidelines

Per clig.dev conventions:

1. **stderr for errors**: All error messages go to stderr, not stdout
2. **Human-readable first**: Default output is for humans; use `--json` for machines
3. **Actionable hints**: Include suggestions for fixing the error when possible
4. **Exit code at end**: In verbose mode, print exit code at end of error output

### Human-Readable Format (default)

```
<Category> error: <brief message>

<Details if applicable>

Hint: <actionable suggestion>
```

### JSON Format (--json flag)

When `--json` flag is present, errors output as JSON to stderr:

```json
{
  "error": {
    "category": "VALIDATION_ERROR",
    "code": 7,
    "message": "Invalid customer ID format",
    "details": "Customer ID must be 10 digits without hyphens",
    "hint": "Use '1234567890' not '123-456-7890'",
    "field": "customer_id",
    "value": "123-456-7890"
  }
}
```

## Go Implementation

```go
package exitcode

// Exit codes per clig.dev conventions
const (
    Success         = 0
    GeneralError    = 1
    UsageError      = 2
    AuthError       = 3
    APIError        = 4
    ConfigError     = 5
    IOError         = 6
    ValidationError = 7
)

// Category returns the error category name for an exit code
func Category(code int) string {
    switch code {
    case Success:
        return "SUCCESS"
    case GeneralError:
        return "GENERAL_ERROR"
    case UsageError:
        return "USAGE_ERROR"
    case AuthError:
        return "AUTH_ERROR"
    case APIError:
        return "API_ERROR"
    case ConfigError:
        return "CONFIG_ERROR"
    case IOError:
        return "IO_ERROR"
    case ValidationError:
        return "VALIDATION_ERROR"
    default:
        return "UNKNOWN"
    }
}
```

## Decision Tree

```
Error occurred?
├── No → Exit 0 (SUCCESS)
└── Yes
    ├── Invalid CLI arguments/flags?
    │   └── Yes → Exit 2 (USAGE_ERROR)
    ├── Config missing or invalid?
    │   └── Yes → Exit 5 (CONFIG_ERROR)
    ├── Input validation failed (before API call)?
    │   └── Yes → Exit 7 (VALIDATION_ERROR)
    ├── Auth/credentials issue?
    │   └── Yes → Exit 3 (AUTH_ERROR)
    ├── API returned error?
    │   └── Yes → Exit 4 (API_ERROR)
    ├── File/network I/O failed?
    │   └── Yes → Exit 6 (IO_ERROR)
    └── Otherwise → Exit 1 (GENERAL_ERROR)
```

## Signal Handling

Standard Unix signal behavior:

| Signal | Exit Code | Behavior |
|--------|-----------|----------|
| SIGINT (Ctrl+C) | 130 | Graceful shutdown, print newline |
| SIGTERM | 143 | Graceful shutdown |
| SIGPIPE | 0 | Silent exit (expected for pipes) |

## Testing

Exit codes should be tested for all error paths:

```bash
# Test usage error
adtap query --badflg 2>/dev/null; echo $?  # Should output: 2

# Test validation error
adtap query "123-456-7890" "SELECT ..." 2>/dev/null; echo $?  # Should output: 7

# Test config error (no config)
HOME=/nonexistent adtap doctor 2>/dev/null; echo $?  # Should output: 5
```

## References

- [clig.dev - Exit Codes](https://clig.dev/#exit-codes)
- [Bash Exit Codes](https://www.gnu.org/software/bash/manual/html_node/Exit-Status.html)
- [sysexits.h](https://man.freebsd.org/cgi/man.cgi?query=sysexits) - BSD exit code conventions
