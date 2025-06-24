package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

type WebsiteStatus struct {
	URL      string
	Status   int
	Duration time.Duration
	IsUp     bool
}

// 清空終端屏幕
func clearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// 調整字符串長度函數
func padString(s string, length int) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat(" ", length-len(s))
}

// 檢查網站並返回狀態
func checkWebsite(url string) WebsiteStatus {
	startTime := time.Now()

	// 確保 URL 有 http/https 前綴
	if len(url) > 4 && url[:4] != "http" {
		url = "https://" + url
	} else if len(url) <= 4 {
		url = "https://" + url
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	duration := time.Since(startTime)

	if err != nil {
		return WebsiteStatus{
			URL:      url,
			Status:   0,
			Duration: duration,
			IsUp:     false,
		}
	}
	defer resp.Body.Close()

	return WebsiteStatus{
		URL:      url,
		Status:   resp.StatusCode,
		Duration: duration,
		IsUp:     resp.StatusCode >= 200 && resp.StatusCode < 400,
	}
}

func main() {
	// 預設網站列表
	websites := []string{
		"google.com",
		"github.com",
		"example.com",
		"joball.tw",
	}

	// 從命令列參數添加網站
	if len(os.Args) > 1 {
		websites = append(websites, os.Args[1:]...)
	}

	// 網站狀態map
	statuses := make(map[string]*WebsiteStatus)

	// 設定欄位寬度
	urlWidth := 25   // 網站欄位寬度
	stateWidth := 10 // 狀態欄位寬度
	timeWidth := 15  // 響應時間欄位寬度
	codeWidth := 15  // 狀態碼欄位寬度

	// 主循環
	for {
		// 檢查所有網站
		var wg sync.WaitGroup
		var mu sync.Mutex

		for _, site := range websites {
			wg.Add(1)
			go func(url string) {
				defer wg.Done()
				status := checkWebsite(url)

				mu.Lock()
				statuses[url] = &status // 直接替換，不保留歷史記錄
				mu.Unlock()
			}(site)
		}

		wg.Wait()

		// 計算統計數據
		total := len(statuses)
		online := 0
		for _, status := range statuses {
			if status.IsUp {
				online++
			}
		}

		// 清屏
		clearScreen()

		// 顯示標題和時間
		fmt.Println(color.CyanString("網站健康監控系統"), color.YellowString(time.Now().Format("2006-01-02 15:04:05")))
		fmt.Println()

		// 顯示總統計信息
		fmt.Printf("總網站數: %-2d 在線: %-2s 離線: %-2s 正常率: %s\n\n",
			total,
			color.GreenString("%d", online),
			color.RedString("%d", total-online),
			color.CyanString("%.1f%%", float64(online)/float64(total)*100),
		)

		// 網站狀態表格 - 非常簡單的格式
		fmt.Println(" 網站監控狀態" + strings.Repeat(" ", 63))
		fmt.Println(strings.Repeat("-", 80))

		// 打印表頭並確保每個欄位都填充到適當的寬度
		fmt.Printf(" %s %s %s %s\n",
			padString("網站", urlWidth),
			padString("狀態", stateWidth),
			padString("響應時間", timeWidth),
			padString("最近檢查", codeWidth),
		)
		fmt.Println(strings.Repeat("-", 80))

		// 顯示每個網站的狀態
		for _, site := range websites {
			status, exists := statuses[site]
			if !exists {
				continue
			}

			// 狀態文字
			statusText := color.RedString("離線")
			if status.IsUp {
				statusText = color.GreenString("在線")
			}

			// 狀態碼
			codeText := ""
			if status.Status > 0 {
				if status.Status >= 200 && status.Status < 400 {
					codeText = color.GreenString("%d", status.Status)
				} else {
					codeText = color.RedString("%d", status.Status)
				}
			} else {
				codeText = color.RedString("ERROR")
			}

			// URL可能包含顏色代碼，先取得染色後的URL
			coloredURL := color.CyanString(status.URL)

			// 顯示行，所有欄位都確保填充到適當寬度
			fmt.Printf(" %s %s %s %s\n",
				padString(coloredURL, urlWidth),
				padString(statusText, stateWidth),
				padString(fmt.Sprintf("%dms", status.Duration.Milliseconds()), timeWidth),
				padString(codeText, codeWidth),
			)
		}

		fmt.Println(strings.Repeat("-", 80))
		fmt.Println("\n按 Ctrl+C 退出監控。每10秒自動更新...")

		// 等待10秒
		time.Sleep(10 * time.Second)
	}
}
