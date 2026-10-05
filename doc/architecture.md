# go-web-monitor - Architecture

> Back to [README](../README.md)

## Overview

```mermaid
graph TB
    Main[cmd/tui/main.go] -->|util.Read / util.Save| Store[(.webMonitor.json)]
    Main -->|service.New| Monitor[Monitor]
    Main -->|service.NewUI| UI[UI]
    UI -->|Ticker every minute / refresh| Monitor
    UI -->|add / del / smtp / test| Monitor
    Monitor -->|Check: HEAD → GET| Site[Target site]
    Monitor -->|util.CheckExpire| TLS[Target site :443]
    Monitor -->|sendEmail / TestEmail| SMTP[SMTP]
    SMTP --> Mail[Mail server]
    Monitor -->|Save| Store
```

## Module: cmd/tui

The entry point initializes the config file and wires Monitor to UI.

```mermaid
graph TB
    subgraph cmd/tui
        Start[main] --> Exists{.webMonitor.json exists?}
        Exists -->|No| Init[Create default MonitorData<br/>port 587 / enabled false]
        Init --> Read
        Exists -->|Yes| Read[util.Read]
        Read --> List[Collect URLs from list]
        List --> NewMon[service.New]
        NewMon --> NewUI[service.NewUI]
        NewUI --> Run[UI.Run]
    end
    Init -->|util.Save| Store[(.webMonitor.json)]
    Read --> Store
    Run -->|error| Exit[os.Exit 1]
```

## Module: service/Monitor

The monitoring core holds the watch list, status table, and SMTP config, guarding shared state with `sync.RWMutex`.

```mermaid
graph TB
    subgraph Monitor
        State[list / status / config / isShow]
        Update[Update] -->|one goroutine per site| Check[Check]
        Update --> Counter[Failure counter]
        Counter -->|5th consecutive offline| SendEmail[sendEmail]
        Update --> Save[Save]
        Add[Add] --> Check
        Remove[Remove] --> State
        UpdateSMTP[UpdateSMTP / AddEmail / RemoveEmail] --> State
        TestEmail[TestEmail]
        Check --> Normalize[Prepend https://]
        Normalize --> Head[HEAD request]
        Head -->|failure or 404| Get[GET request]
        Counter --> State
    end
    Head --> Site[Target site]
    Get --> Site
    Check -->|request succeeded| CheckExpire[util.CheckExpire]
    SendEmail --> SMTP[SMTP.SendEmail]
    TestEmail --> SMTP
    UpdateSMTP -->|UpdateConfig| SMTP
    Save -->|util.Save| Store[(.webMonitor.json)]
```

## Module: service/UI

The tview-based terminal interface schedules checks, renders the table, and parses commands.

```mermaid
graph TB
    subgraph UI
        Run[Run] --> Setup[setupUI]
        Run --> Loop[start: 1-minute Ticker]
        Setup --> Table[Monitor table]
        Setup --> Text[System panel]
        Setup --> Input[Command input]
        Input -->|Enter| Handle[handleCommand]
        Loop -->|QueueUpdateDraw| Render[updateUI]
        Render --> Table
        Render --> Info[updateInfo]
        Info -->|isShow = false| Stats[Stats and command help]
        Info -->|isShow = true| ShowSMTP[showSMTP]
        Stats --> Text
        ShowSMTP --> Text
        Keys[Ctrl+C / Esc] --> Capture[SetInputCapture]
    end
    Loop --> Update[Monitor.Update]
    Handle --> Mon[Monitor command methods]
    Capture -->|close quit| Quit[Monitor.GetQuit]
    Quit --> Loop
```

## Module: service/SMTP

Picks the connection mode by port and sends HTML email.

```mermaid
graph TB
    subgraph SMTP
        Send[SendEmail] --> Guard{enabled, host set,<br/>and to non-empty?}
        Guard -->|No| Err[Return error]
        Guard -->|Yes| Probe[TCP probe 5s]
        Probe --> Plain[smtp.Dial plain]
        Plain -->|fails on port 465| Implicit[tls.Dial implicit TLS]
        Plain -->|succeeds on port 587| StartTLS[STARTTLS]
        Plain -->|succeeds| Auth
        Implicit --> Auth{username and password both set?}
        StartTLS --> Auth
        Auth -->|Yes| Plainauth[PLAIN auth]
        Auth -->|No| Mail
        Plainauth --> Mail[MAIL FROM / RCPT TO]
        Mail --> Data[Write HTML message<br/>cc falls back to from]
    end
    Data --> Server[Mail server]
```

## Module: internal/util

Stateless helpers for config file IO and SSL expiry calculation.

```mermaid
graph LR
    subgraph util
        Save[Save] -->|json.MarshalIndent| Write[os.WriteFile 0644]
        Read[Read] -->|os.ReadFile| Parse[json.Unmarshal]
        CheckExpire[CheckExpire] --> Host[Extract host]
        Host --> Dial[tls.Dial host:443]
        Dial --> Cert[PeerCertificates 0 .NotAfter]
        Cert --> Days[Days remaining]
    end
    Write --> Store[(.webMonitor.json)]
    Store --> Read
```

## Module: internal/model

```mermaid
classDiagram
    class MonitorData {
        +[]Website List
        +Config Config
    }
    class Website {
        +string URL
        +int Code
        +time.Duration Duration
        +bool Online
        +time.Time LastCheck
        +int Expire
        +int Count
    }
    class Config {
        +string Host
        +int Port
        +string Username
        +string Password
        +string From
        +[]string To
        +string CC
        +bool Enabled
    }
    MonitorData --> Website
    MonitorData --> Config
```

## Data Flow

```mermaid
sequenceDiagram
    participant T as Ticker
    participant U as UI
    participant M as Monitor
    participant W as Target site
    participant S as SMTP
    participant F as .webMonitor.json
    T->>U: Fires every minute
    U->>M: Update()
    par Each site concurrently
        M->>W: HEAD
        alt Failure or 404
            M->>W: GET
        end
        W-->>M: Status code
        M->>W: TLS :443 read certificate
        W-->>M: NotAfter
        alt 5th consecutive offline
            M-)S: sendEmail (async)
        end
    end
    M->>F: Save()
    U->>U: QueueUpdateDraw(updateUI)
```

## State Machine

### Site Failure Counter

```mermaid
stateDiagram-v2
    [*] --> Healthy
    Healthy: Online (count = 0)
    Failing: Offline (count 1–4)
    Healthy --> Failing: Offline
    Failing --> Failing: Offline and count < 4
    Failing --> Healthy: Online
    Failing --> Alerted: 5th offline
    Alerted: Alert sent, count reset
    Alerted --> Failing: Still offline
    Alerted --> Healthy: Online
```

### UI Views

```mermaid
stateDiagram-v2
    [*] --> Main
    Main: Main view (stats)
    SMTPView: SMTP settings view
    Main --> SMTPView: smtp / smtp show
    SMTPView --> Main: back / Esc
    Main --> [*]: quit / Ctrl+C
    SMTPView --> [*]: quit / Ctrl+C
```

***

©️ 2025 [邱敬幃 Pardn Chiu](https://www.linkedin.com/in/pardnchiu)
