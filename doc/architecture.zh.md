# go-web-monitor - 架構

> 返回 [README](./README.zh.md)

## 概覽

```mermaid
graph TB
    Main[cmd/tui/main.go] -->|util.Read / util.Save| Store[(.webMonitor.json)]
    Main -->|service.New| Monitor[Monitor]
    Main -->|service.NewUI| UI[UI]
    UI -->|Ticker 每分鐘 / refresh| Monitor
    UI -->|add / del / smtp / test| Monitor
    Monitor -->|Check: HEAD → GET| Site[目標網站]
    Monitor -->|util.CheckExpire| TLS[目標網站 :443]
    Monitor -->|sendEmail / TestEmail| SMTP[SMTP]
    SMTP --> Mail[郵件伺服器]
    Monitor -->|Save| Store
```

## Module: cmd/tui

程式入口，負責初始化設定檔並組裝 Monitor 與 UI。

```mermaid
graph TB
    subgraph cmd/tui
        Start[main] --> Exists{.webMonitor.json 存在?}
        Exists -->|否| Init[建立預設 MonitorData<br/>port 587 / enabled false]
        Init --> Read
        Exists -->|是| Read[util.Read]
        Read --> List[取出 list 的 URL]
        List --> NewMon[service.New]
        NewMon --> NewUI[service.NewUI]
        NewUI --> Run[UI.Run]
    end
    Init -->|util.Save| Store[(.webMonitor.json)]
    Read --> Store
    Run -->|錯誤| Exit[os.Exit 1]
```

## Module: service/Monitor

監控核心，持有監控清單、狀態表與 SMTP 設定，以 `sync.RWMutex` 保護共享狀態。

```mermaid
graph TB
    subgraph Monitor
        State[list / status / config / isShow]
        Update[Update] -->|每站一個 goroutine| Check[Check]
        Update --> Counter[失敗計數]
        Counter -->|連續第 5 次離線| SendEmail[sendEmail]
        Update --> Save[Save]
        Add[Add] --> Check
        Remove[Remove] --> State
        UpdateSMTP[UpdateSMTP / AddEmail / RemoveEmail] --> State
        TestEmail[TestEmail]
        Check --> Normalize[補 https://]
        Normalize --> Head[HEAD 請求]
        Head -->|失敗或 404| Get[GET 請求]
        Counter --> State
    end
    Head --> Site[目標網站]
    Get --> Site
    Check -->|請求成功| CheckExpire[util.CheckExpire]
    SendEmail --> SMTP[SMTP.SendEmail]
    TestEmail --> SMTP
    UpdateSMTP -->|UpdateConfig| SMTP
    Save -->|util.Save| Store[(.webMonitor.json)]
```

## Module: service/UI

基於 tview 的終端介面，負責排程檢查、渲染表格與解析指令。

```mermaid
graph TB
    subgraph UI
        Run[Run] --> Setup[setupUI]
        Run --> Loop[start: Ticker 1 分鐘]
        Setup --> Table[Monitor 表格]
        Setup --> Text[System 面板]
        Setup --> Input[Command 輸入框]
        Input -->|Enter| Handle[handleCommand]
        Loop -->|QueueUpdateDraw| Render[updateUI]
        Render --> Table
        Render --> Info[updateInfo]
        Info -->|isShow = false| Stats[統計與指令說明]
        Info -->|isShow = true| ShowSMTP[showSMTP]
        Stats --> Text
        ShowSMTP --> Text
        Keys[Ctrl+C / Esc] --> Capture[SetInputCapture]
    end
    Loop --> Update[Monitor.Update]
    Handle --> Mon[Monitor 指令方法]
    Capture -->|close quit| Quit[Monitor.GetQuit]
    Quit --> Loop
```

## Module: service/SMTP

依 port 選擇連線方式並以 HTML 格式寄信。

```mermaid
graph TB
    subgraph SMTP
        Send[SendEmail] --> Guard{enabled 且 host 已設<br/>且 to 非空?}
        Guard -->|否| Err[回傳錯誤]
        Guard -->|是| Probe[TCP 探測 5 秒]
        Probe --> Plain[smtp.Dial 明文]
        Plain -->|失敗且 port 465| Implicit[tls.Dial 隱式 TLS]
        Plain -->|成功且 port 587| StartTLS[STARTTLS]
        Plain -->|成功| Auth
        Implicit --> Auth{username 與 password 皆非空?}
        StartTLS --> Auth
        Auth -->|是| Plainauth[PLAIN 認證]
        Auth -->|否| Mail
        Plainauth --> Mail[MAIL FROM / RCPT TO]
        Mail --> Data[寫入 HTML 信件<br/>cc 空值時以 from 代替]
    end
    Data --> Server[郵件伺服器]
```

## Module: internal/util

無狀態工具：設定檔讀寫與 SSL 憑證到期天數計算。

```mermaid
graph LR
    subgraph util
        Save[Save] -->|json.MarshalIndent| Write[os.WriteFile 0644]
        Read[Read] -->|os.ReadFile| Parse[json.Unmarshal]
        CheckExpire[CheckExpire] --> Host[擷取 host]
        Host --> Dial[tls.Dial host:443]
        Dial --> Cert[PeerCertificates 0 .NotAfter]
        Cert --> Days[剩餘天數]
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

## 資料流

```mermaid
sequenceDiagram
    participant T as Ticker
    participant U as UI
    participant M as Monitor
    participant W as 目標網站
    participant S as SMTP
    participant F as .webMonitor.json
    T->>U: 每分鐘觸發
    U->>M: Update()
    par 每個網站並行
        M->>W: HEAD
        alt 失敗或 404
            M->>W: GET
        end
        W-->>M: 狀態碼
        M->>W: TLS :443 讀取憑證
        W-->>M: NotAfter
        alt 連續第 5 次離線
            M-)S: sendEmail（非同步）
        end
    end
    M->>F: Save()
    U->>U: QueueUpdateDraw(updateUI)
```

## 狀態機

### 網站失敗計數

```mermaid
stateDiagram-v2
    [*] --> Healthy
    Healthy: 上線（count = 0）
    Failing: 離線（count 1–4）
    Healthy --> Failing: 離線
    Failing --> Failing: 離線且 count < 4
    Failing --> Healthy: 上線
    Failing --> Alerted: 第 5 次離線
    Alerted: 寄出告警，count 歸零
    Alerted --> Failing: 仍離線
    Alerted --> Healthy: 上線
```

### 介面畫面

```mermaid
stateDiagram-v2
    [*] --> Main
    Main: 主畫面（統計）
    SMTPView: SMTP 設定畫面
    Main --> SMTPView: smtp / smtp show
    SMTPView --> Main: back / Esc
    Main --> [*]: quit / Ctrl+C
    SMTPView --> [*]: quit / Ctrl+C
```

***

©️ 2025 [邱敬幃 Pardn Chiu](https://www.linkedin.com/in/pardnchiu)
