# truelist-cli

The official command-line tool for [Truelist.io](https://truelist.io) email validation. Validate single emails, bulk CSV files, or pipe from stdin.

<!-- TODO: Add animated demo GIF -->
<!-- ![truelist-cli demo](docs/demo.gif) -->

## Installation

### Homebrew (macOS/Linux)

```bash
brew install Truelist-io-Email-Validation/tap/truelist
```

### Go install

```bash
go install github.com/Truelist-io-Email-Validation/truelist-cli@latest
```

### Binary download

Download the latest release from the [Releases page](https://github.com/Truelist-io-Email-Validation/truelist-cli/releases) and add it to your `PATH`.

## Quick Start

```bash
# Set your API key (get one at https://truelist.io)
truelist config set api-key YOUR_API_KEY

# Validate a single email
truelist validate user@example.com

# Validate a CSV file
truelist validate --file contacts.csv

# Pipe emails from stdin
cat emails.txt | truelist validate
```

## Commands

### `truelist validate <email>`

Validate a single email address.

```bash
truelist validate user@gmail.com
```

```
✓ user@gmail.com
  State:       ok
  Sub-state:   email_ok
  Domain:      gmail.com
  Canonical:   user
```

**Flags:**
| Flag | Description |
|------|-------------|
| `--json` | Output result as JSON |
| `-q, --quiet` | Output only the state (`ok`, `email_invalid`, `accept_all`) |

### `truelist validate --file <path>`

Validate emails from a CSV file. The CLI auto-detects the email column, or you can specify it with `--column`.

```bash
truelist validate --file contacts.csv
truelist validate --file contacts.csv --column email_address
truelist validate --file contacts.csv --output results.csv
```

Outputs a new CSV with `truelist_state`, `truelist_sub_state`, `truelist_domain`, `truelist_verified_at`, and `truelist_suggestion` columns appended.

**Flags:**
| Flag | Description |
|------|-------------|
| `-f, --file` | Path to the input CSV file |
| `-o, --output` | Output file path (default: `<input>_validated.csv`) |
| `-c, --column` | Name of the email column in the CSV |

### `truelist validate` (stdin)

Pipe emails from stdin, one per line.

```bash
cat emails.txt | truelist validate
echo "user@example.com" | truelist validate
```

### `truelist whoami`

Check your API key and display account information.

```bash
truelist whoami
```

```
Account Info
  Email:      you@company.com
  Name:       Your Name
  UUID:       abc-123
  Time Zone:  America/New_York
  Plan:       pro
```

### `truelist config set api-key <key>`

Save your API key to the config file.

```bash
truelist config set api-key tk_live_abc123
```

### `truelist version`

Print the CLI version.

```bash
truelist version
```

## Configuration

The CLI looks for your API key in the following order:

1. **Config file** at `~/.config/truelist/config.yaml`
2. **Environment variable** `TRUELIST_API_KEY`

Set via config file:

```bash
truelist config set api-key YOUR_API_KEY
```

Set via environment variable:

```bash
export TRUELIST_API_KEY=YOUR_API_KEY
```

## Output Formats

### Human-readable (default)

```
✓ user@gmail.com
  State:       ok
  Sub-state:   email_ok
  Domain:      gmail.com
  Canonical:   user
  Verified At: 2026-02-21T10:00:00.000Z
```

### JSON (`--json`)

```json
{
  "address": "user@gmail.com",
  "domain": "gmail.com",
  "canonical": "user",
  "mx_record": null,
  "first_name": null,
  "last_name": null,
  "email_state": "ok",
  "email_sub_state": "email_ok",
  "verified_at": "2026-02-21T10:00:00.000Z",
  "did_you_mean": null
}
```

### Quiet (`--quiet`)

```
ok
```

## Validation States

| State | Description |
|-------|-------------|
| `ok` | The email is deliverable |
| `email_invalid` | The email is not deliverable |
| `accept_all` | Domain accepts all addresses (catch-all) |

### Sub-states

| Sub-state | Description |
|-----------|-------------|
| `email_ok` | Email is valid and deliverable |
| `is_disposable` | Temporary/disposable email address |
| `is_role` | Role-based address (e.g., info@, admin@) |
| `failed_mx_check` | Domain has no valid MX records |
| `failed_spam_trap` | Address is a known spam trap |
| `failed_no_mailbox` | Mailbox does not exist |
| `failed_greylisted` | Server temporarily rejected the request |
| `failed_syntax_check` | Email address has invalid syntax |

## Rate Limits

The CLI respects Truelist API rate limits (10 requests/second). Bulk validation automatically throttles requests.

## Development

```bash
# Build
make build

# Run tests
make test

# Build for all platforms
make build-all
```

## License

MIT - see [LICENSE](LICENSE) for details.
