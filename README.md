> [!NOTE]
> This README was translated by ChatGPT 4o

# Website Monitoring System

> Lightweight Go-based website monitoring tool offering real-time status checks, SSL certificate monitoring, and email notifications. Includes a TUI interface for convenient status management.

![lang](https://img.shields.io/badge/lang-Go-blue)
[![license](https://img.shields.io/github/license/pardnchiu/web-monitor)](LICENSE)
[![version](https://img.shields.io/github/v/tag/pardnchiu/web-monitor)](https://github.com/pardnchiu/web-monitor/releases)<br>
[![readme](https://img.shields.io/badge/readme-EN-white)](README.md)
[![readme](https://img.shields.io/badge/readme-ZH-white)](README.zh.md)

## Key Features

### Website Monitoring
- HTTP/HTTPS status code checks
- SSL certificate expiration monitoring
- Response time measurement
- Automatic retry mechanism
- Multi-site monitoring support

### Email Notifications
- SMTP email sending
- SSL/TLS connection support (ports 465/587)
- Customizable email content
- Multi-recipient support

### TUI Interactive Interface
- Real-time status display
- Command-line operations
- Dynamic site addition/removal
- SMTP configuration management

## Dependencies

- [`github.com/gdamore/tcell/v2`](https://github.com/gdamore/tcell) - Terminal interface
- [`github.com/rivo/tview`](https://github.com/rivo/tview) - TUI component library

## Usage

### Installation

```bash
# Clone the repository
git clone https://github.com/pardnchiu/web-monitor.git
cd web-monitor

# Compile and run
go run cmd/tui/main.go
```

### Basic Operations

```bash
# Add a website for monitoring
add example.com

# Remove a monitored website  
del example.com

# Configure SMTP
smtp host smtp.gmail.com
smtp port 587
smtp username your@email.com
smtp password your_password
smtp from your@email.com
smtp add recipient@email.com
smtp enabled true

# Test email sending
test

# Manually refresh monitoring
refresh

# Exit the program
quit
```

## Configuration File Format

The system automatically generates a `.webMonitor.json` configuration file:

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
    "password": "your_password",
    "from": "your@email.com",
    "to": ["recipient@email.com"],
    "cc": "your@email.com",
    "enabled": true
  }
}
```

## Monitoring Mechanism

### Check Frequency
- All websites are checked every minute.
- Concurrent processing improves efficiency.
- Results are automatically saved.

### Notification Rules
- Email notifications are triggered after 5 consecutive failures.
- Includes failure reasons and timestamps.
- HTML-formatted email content.

### SSL Monitoring
- Automatically checks HTTPS website certificates.
- Displays remaining expiration days.
- Color-coded indicators (Green: >30 days, Yellow: 7-30 days, Red: <7 days).

## TUI Interface

### Main Screen
- Website status displayed in a table.
- Real-time monitoring updates.
- System statistics.

## License

This project is licensed under the [MIT](LICENSE) license.

## Author

<img src="https://avatars.githubusercontent.com/u/25631760" align="left" width="96" height="96" style="margin-right: 0.5rem;">

<h4 style="padding-top: 0">邱敬幃 Pardn Chiu</h4>

<a href="mailto:dev@pardn.io" target="_blank">
  <img src="https://pardn.io/image/email.svg" width="48" height="48">
</a> <a href="https://linkedin.com/in/pardnchiu" target="_blank">
  <img src="https://pardn.io/image/linkedin.svg" width="48" height="48">
</a>

***

©️ 2025 [邱敬幃 Pardn Chiu](https://pardn.io)