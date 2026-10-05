# go-web-monitor - Documentation

> Back to [README](../README.md)

## Prerequisites

- Go 1.21 or higher
- A terminal with ANSI color support
- Outbound access to port 443 (SSL certificate check)
- An SMTP account (optional, required only for email alerts)

## Installation

### Run from Source

```bash
git clone https://github.com/pardnchiu/go-web-monitor.git
cd go-web-monitor
go run ./cmd/tui
```

### Build from Source

```bash
git clone https://github.com/pardnchiu/go-web-monitor.git
cd go-web-monitor
go build -o go-web-monitor ./cmd/tui
./go-web-monitor
```

> The module name in `go.mod` is `website-monitor`, which does not match the GitHub path, so `go install github.com/pardnchiu/go-web-monitor/...` does not work.

## Configuration

The tool reads no environment variables. All state and settings live in `.webMonitor.json` in the **current working directory** (already listed in `.gitignore`).

| When | Behavior |
|------|----------|
| File missing at startup | Creates an empty list with SMTP port `587` and `enabled: false` |
| At startup | Loads `list[].url` as the watch list and `config` as the SMTP settings |
| After each check round and after `add` / `del` / `smtp` commands | Overwrites the whole file with the in-memory state |

### Config File Structure

```json
{
  "list": [
    {
      "url": "https://example.com",
      "code": 200,
      "duration": 62554917,
      "online": true,
      "last_check": "2025-01-15T23:12:29.200334+08:00",
      "expire": 87,
      "count": 0
    }
  ],
  "config": {
    "host": "smtp.gmail.com",
    "port": 587,
    "username": "your@email.com",
    "password": "your_app_password",
    "from": "your@email.com",
    "to": ["recipient@email.com"],
    "cc": "",
    "enabled": true
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `list[].url` | string | Normalized URL (`https://` is prepended when the input does not start with `http`) |
| `list[].code` | int | Last HTTP status code; `0` on connection failure |
| `list[].duration` | int | Response time in nanoseconds |
| `list[].online` | bool | `true` when the status code is 200–399 |
| `list[].last_check` | string | Last check time (RFC 3339) |
| `list[].expire` | int | Days until the SSL certificate expires; `0` when unavailable |
| `list[].count` | int | Consecutive failure count (0–4) |
| `config.host` | string | SMTP host |
| `config.port` | int | SMTP port; `465` uses implicit TLS, `587` uses STARTTLS |
| `config.username` / `config.password` | string | PLAIN auth runs only when both are non-empty |
| `config.from` | string | Sender address |
| `config.to` | []string | Recipient list |
| `config.cc` | string | CC address; falls back to `from` when empty. No TUI command sets it, so edit it by hand |
| `config.enabled` | bool | Enables email alerts |

> The program overwrites the config file while running, so make manual edits (such as `cc`) only after it exits.

## Usage

### Basic

After launch, the cursor sits in the Command input at the bottom right. Type a command and press Enter:

```bash
# Add a site (https:// is prepended automatically)
add example.com
add https://api.example.com/health

# Re-check all sites now
refresh

# Remove a site
del example.com

# Exit
quit
```

The Monitor table on the left shows `Website`, `Status`, `Duration`, `Code`, `Last Check`, and `SSL Expire`. The System panel on the right shows the total, online and offline counts, uptime, and email alert status.

### Email Alerts (Gmail, 587 STARTTLS)

```bash
smtp host smtp.gmail.com
smtp port 587
smtp username your@gmail.com
smtp password abcdefghijklmnop
smtp from your@gmail.com
smtp add ops@example.com
smtp add oncall@example.com
smtp enabled true

# Show the current SMTP settings (password masked with *)
smtp

# Send a test email
test

# Return to the main screen (or press Esc)
back
```

Commands split on whitespace and each setting takes only the first token after the key. Gmail displays app passwords with spaces; remove them when entering the value.

### Implicit TLS on 465

```bash
smtp host smtp.example.com
smtp port 465
smtp enabled 1
test
```

Connection flow: probe TCP with a 5-second timeout, then try plain SMTP. When plain SMTP fails on port `465`, connect over TLS directly; on port `587`, upgrade with STARTTLS when the server supports it.

### Alert Rules

| Situation | Behavior |
|-----------|----------|
| Check result is online | Reset the failure count |
| Offline with count below 5 | Increment the count |
| 5th consecutive offline check | Send `Website Offline Alert: <url>` and reset the count |
| Still offline | Send again every 5 failures (about every 5 minutes) |

### Notes

- After a restart, the watch list loads the normalized URLs (with `https://`) from the config file, so removal needs the full URL, e.g. `del https://example.com`.
- `add` runs one check synchronously; an unresponsive target can block for up to 10 seconds each for HEAD and GET.
- Errors from `add` / `del` / `smtp` and `test` failures do not appear in the System panel.

## CLI Reference

### Launch

| Command | Description |
|---------|-------------|
| `go run ./cmd/tui` / `./go-web-monitor` | Starts the TUI; takes no flags or arguments |

### Main Commands

| Command | Syntax | Description |
|---------|--------|-------------|
| `add` | `add <url>` | Adds a site and checks it once; duplicates are rejected |
| `delete` / `del` / `remove` / `rm` | `del <url>` | Removes a site |
| `smtp` / `smtp show` | `smtp` | Switches to the SMTP settings view |
| `smtp` | `smtp <key> <value>` | Sets an SMTP field (see below) |
| `test` | `test` | Sends a test email in the background |
| `refresh` | `refresh` | Re-checks all sites now |
| `back` | `back` | Returns from the SMTP view to the main view |
| `quit` / `exit` / `bye` | `quit` | Exits the program |

### SMTP Keys

| Key | Value | Description |
|-----|-------|-------------|
| `host` | Hostname | SMTP host |
| `port` | Integer | Rejected when not an integer |
| `username` | String | Auth username |
| `password` | String | Auth password |
| `from` | Email | Sender address |
| `enabled` | `true` / `1` / other | `true` or `1` enables; anything else disables |
| `add` | Email | Adds a recipient (duplicates ignored) |
| `delete` / `del` / `remove` / `rm` | Email | Removes a recipient |

### Key Bindings

| Key | Description |
|-----|-------------|
| `Enter` | Submit the command |
| `Esc` | Return from the SMTP view to the main view |
| `Ctrl+C` | Exit the program |

### Probe Behavior

| Item | Value |
|------|-------|
| Check interval | Every minute, all sites concurrently |
| Request timeout | 10 seconds (HEAD and GET counted separately) |
| Request order | HEAD; GET on failure or 404 |
| Redirects | Followed |
| Online criteria | Status code 200–399 |
| SSL check | After a successful request, dials TLS to `host:443` and reads the certificate |
| SSL colors | >30 days green, 8–30 days yellow, ≤7 days or unavailable red |

***

©️ 2025 [邱敬幃 Pardn Chiu](https://www.linkedin.com/in/pardnchiu)
