# go-web-monitor - 技術文件

> 返回 [README](./README.zh.md)

## 前置需求

- Go 1.21 或更高版本
- 支援 ANSI 色彩的終端機
- 可對外連線的 443 埠（SSL 憑證檢查）
- SMTP 伺服器帳號（選用，僅 Email 告警需要）

## 安裝

### 從原始碼執行

```bash
git clone https://github.com/pardnchiu/go-web-monitor.git
cd go-web-monitor
go run ./cmd/tui
```

### 從原始碼建置

```bash
git clone https://github.com/pardnchiu/go-web-monitor.git
cd go-web-monitor
go build -o go-web-monitor ./cmd/tui
./go-web-monitor
```

> `go.mod` 的 module 名稱為 `website-monitor`，與 GitHub 路徑不一致，因此無法使用 `go install github.com/pardnchiu/go-web-monitor/...` 安裝。

## 設定

本工具不讀取環境變數，所有狀態與設定都存放於**當前工作目錄**下的 `.webMonitor.json`（已列入 `.gitignore`）。

| 時機 | 行為 |
|------|------|
| 啟動時檔案不存在 | 自動建立空白清單，SMTP 預設 port `587`、`enabled: false` |
| 啟動時 | 讀取 `list[].url` 作為監控清單、`config` 作為 SMTP 設定 |
| 每輪檢查完成、`add`／`del`／`smtp` 指令後 | 以記憶體中的狀態整份覆寫 |

### 設定檔結構

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

| 欄位 | 型別 | 說明 |
|------|------|------|
| `list[].url` | string | 正規化後的網址（未以 `http` 開頭者自動補 `https://`） |
| `list[].code` | int | 最近一次的 HTTP 狀態碼，連線失敗為 `0` |
| `list[].duration` | int | 回應耗時（奈秒） |
| `list[].online` | bool | 狀態碼介於 200–399 為 `true` |
| `list[].last_check` | string | 最近一次檢查時間（RFC 3339） |
| `list[].expire` | int | SSL 憑證剩餘天數，無法取得為 `0` |
| `list[].count` | int | 連續失敗次數（0–4） |
| `config.host` | string | SMTP 主機 |
| `config.port` | int | SMTP 埠；`465` 走隱式 TLS、`587` 走 STARTTLS |
| `config.username` / `config.password` | string | 兩者皆非空時才進行 PLAIN 認證 |
| `config.from` | string | 寄件者 |
| `config.to` | []string | 收件者清單 |
| `config.cc` | string | 副本；空值時以 `from` 代替。TUI 無對應指令，僅能手動編輯 |
| `config.enabled` | bool | 是否啟用 Email 告警 |

> 程式執行中會覆寫設定檔，手動編輯（例如 `cc`）請在程式結束後進行。

## 使用方式

### 基礎用法

啟動後游標停在右下角的 Command 輸入框，輸入指令並按 Enter：

```bash
# 新增監控網站（自動補 https://）
add example.com
add https://api.example.com/health

# 立即重新檢查全部網站
refresh

# 移除網站
del example.com

# 離開
quit
```

左側 Monitor 表格欄位：`Website`、`Status`、`Duration`、`Code`、`Last Check`、`SSL Expire`。右側 System 面板顯示總數、上線／離線數、可用率與 Email 告警狀態。

### 設定 Email 告警（Gmail，587 STARTTLS）

```bash
smtp host smtp.gmail.com
smtp port 587
smtp username your@gmail.com
smtp password abcdefghijklmnop
smtp from your@gmail.com
smtp add ops@example.com
smtp add oncall@example.com
smtp enabled true

# 檢視目前 SMTP 設定（密碼以 * 遮蔽）
smtp

# 寄送測試信
test

# 回到主畫面（或按 Esc）
back
```

指令以空白分隔，每個設定值只取第一段；Gmail 應用程式密碼顯示時帶空白，輸入時須去除空白。

### 使用 465 隱式 TLS

```bash
smtp host smtp.example.com
smtp port 465
smtp enabled 1
test
```

連線流程：先以 5 秒逾時探測 TCP，再嘗試明文 SMTP；明文失敗且 port 為 `465` 時改走 TLS 直連，port 為 `587` 且伺服器支援時升級 STARTTLS。

### 告警規則

| 情境 | 行為 |
|------|------|
| 檢查結果上線 | 失敗計數歸零 |
| 離線且計數未達 5 | 計數 +1 |
| 連續第 5 次離線 | 寄出 `Website Offline Alert: <url>`，計數歸零 |
| 持續離線 | 每累積 5 次（約 5 分鐘）再寄一次 |

### 注意事項

- 重新啟動後，監控清單以設定檔中正規化的網址（含 `https://`）載入；此時刪除需輸入完整網址，例如 `del https://example.com`。
- `add` 會同步執行一次檢查，目標無回應時最多等待 HEAD 與 GET 各 10 秒。
- `add`／`del`／`smtp` 的錯誤與 `test` 失敗訊息不會顯示在 System 面板。

## 命令列參考

### 啟動

| 指令 | 說明 |
|------|------|
| `go run ./cmd/tui` / `./go-web-monitor` | 啟動 TUI，無任何旗標與參數 |

### 主畫面指令

| 指令 | 語法 | 說明 |
|------|------|------|
| `add` | `add <url>` | 新增網站並立即檢查一次；重複網址不加入 |
| `delete` / `del` / `remove` / `rm` | `del <url>` | 移除網站 |
| `smtp` / `smtp show` | `smtp` | 切換至 SMTP 設定畫面 |
| `smtp` | `smtp <key> <value>` | 設定 SMTP 欄位（見下表） |
| `test` | `test` | 背景寄送測試信 |
| `refresh` | `refresh` | 立即重新檢查全部網站 |
| `back` | `back` | 從 SMTP 畫面返回主畫面 |
| `quit` / `exit` / `bye` | `quit` | 結束程式 |

### SMTP 設定鍵

| 鍵 | 值 | 說明 |
|----|----|------|
| `host` | 主機名稱 | SMTP 主機 |
| `port` | 整數 | 非整數時拒絕更新 |
| `username` | 字串 | 認證帳號 |
| `password` | 字串 | 認證密碼 |
| `from` | Email | 寄件者 |
| `enabled` | `true` / `1` / 其他 | `true` 或 `1` 為啟用，其餘一律停用 |
| `add` | Email | 新增收件者（重複忽略） |
| `delete` / `del` / `remove` / `rm` | Email | 移除收件者 |

### 快捷鍵

| 按鍵 | 說明 |
|------|------|
| `Enter` | 送出指令 |
| `Esc` | 從 SMTP 畫面返回主畫面 |
| `Ctrl+C` | 結束程式 |

### 探測行為

| 項目 | 值 |
|------|----|
| 檢查週期 | 每 1 分鐘，所有網站並行 |
| 請求逾時 | 10 秒（HEAD、GET 各自計算） |
| 請求順序 | HEAD；失敗或 404 時改送 GET |
| 重新導向 | 跟隨 |
| 上線判定 | 狀態碼 200–399 |
| SSL 檢查 | 請求成功後以 TLS 連線 `host:443` 讀取憑證 |
| SSL 顏色 | >30 天綠、8–30 天黃、≤7 天或無法取得紅 |

***

©️ 2025 [邱敬幃 Pardn Chiu](https://www.linkedin.com/in/pardnchiu)
