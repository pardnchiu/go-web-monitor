# 網站監控系統

> 輕量級的 Go 網站監控工具，提供即時狀態檢查、SSL 憑證監控和電子郵件通知功能。支援 TUI 介面，方便即時查看和管理監控狀態。

![lang](https://img.shields.io/badge/lang-Go-blue)
[![license](https://img.shields.io/github/license/pardnchiu/web-monitor)](LICENSE)
[![version](https://img.shields.io/github/v/tag/pardnchiu/web-monitor)](https://github.com/pardnchiu/web-monitor/releases)<br>
[![readme](https://img.shields.io/badge/readme-EN-white)](README.md)
[![readme](https://img.shields.io/badge/readme-ZH-white)](README.zh.md)

## 三大核心特色

### 網站監控
- HTTP/HTTPS 狀態碼檢查
- SSL 憑證到期日檢查
- 回應時間測量
- 自動重試機制
- 支援多網站同時監控

### 電子郵件通知
- SMTP 郵件發送
- 支援 SSL/TLS 連線 (465/587 埠)
- 自訂郵件內容
- 多收件者支援

### TUI 互動介面
- 即時狀態顯示
- 命令行操作
- 動態新增/移除網站
- SMTP 設定管理

## 依賴套件

- [`github.com/gdamore/tcell/v2`](https://github.com/gdamore/tcell) - 終端機介面
- [`github.com/rivo/tview`](https://github.com/rivo/tview) - TUI 元件庫

## 使用方法

### 安裝

```bash
# 複製專案
git clone https://github.com/pardnchiu/web-monitor.git
cd web-monitor

# 編譯執行
go run cmd/tui/main.go
```

### 基本操作

```bash
# 新增網站監控
add example.com

# 移除網站監控  
del example.com

# 設定 SMTP
smtp host smtp.gmail.com
smtp port 587
smtp username your@email.com
smtp password your_password
smtp from your@email.com
smtp add recipient@email.com
smtp enabled true

# 測試郵件發送
test

# 手動重新整理
refresh

# 結束程式
quit
```

## 設定檔格式

系統會自動建立 `.webMonitor.json` 設定檔：

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

## 監控機制

### 檢查頻率
- 每分鐘檢查一次所有網站
- 使用併發處理提升效率
- 自動儲存檢查結果

### 通知規則
- 連續失敗 5 次觸發郵件通知
- 包含失敗原因和時間資訊
- HTML 格式郵件內容

### SSL 監控
- 自動檢查 HTTPS 網站憑證
- 顯示剩餘到期天數
- 支援顏色標示（綠色：>30 天，黃色：7-30 天，紅色：<7 天）

## TUI 介面

### 主畫面
- 網站狀態表格顯示
- 即時更新監控資訊
- 系統統計資料

## 授權條款

此專案採用 [MIT](LICENSE) 授權條款。

## 作者

<img src="https://avatars.githubusercontent.com/u/25631760" align="left" width="96" height="96" style="margin-right: 0.5rem;">

<h4 style="padding-top: 0">邱敬幃 Pardn Chiu</h4>

<a href="mailto:dev@pardn.io" target="_blank">
  <img src="https://pardn.io/image/email.svg" width="48" height="48">
</a> <a href="https://linkedin.com/in/pardnchiu" target="_blank">
  <img src="https://pardn.io/image/linkedin.svg" width="48" height="48">
</a>

***

©️ 2025 [邱敬幃 Pardn Chiu](https://pardn.io)