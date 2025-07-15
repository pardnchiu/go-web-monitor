package service

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type UI struct {
	app     *tview.Application
	table   *tview.Table
	text    *tview.TextView
	input   *tview.InputField
	monitor *Monitor
}

func NewUI(mon *Monitor) *UI {
	return &UI{
		app:     tview.NewApplication(),
		monitor: mon,
	}
}

func (u *UI) setupUI() {
	u.table = tview.NewTable().
		SetFixed(1, 0).
		SetBorders(true)
	u.table.SetTitle("Monitor").
		SetTitleAlign(tview.AlignLeft).
		SetBorder(true)

	u.text = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false)
	u.text.SetTitle("System").
		SetTitleAlign(tview.AlignLeft).
		SetBorder(true)

	u.input = tview.NewInputField().
		SetFieldWidth(0)
	u.input.SetTitle("Command").
		SetTitleAlign(tview.AlignLeft).
		SetBorder(true)
	u.input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			cmd := u.input.GetText()
			u.input.SetText("")

			if cmd != "" {
				u.handleCommand(cmd)
			}
		}
	})

	rightBlock := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(u.text, 0, 1, false).
		AddItem(u.input, 3, 0, true)

	mainBlock := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(u.table, 0, 2, false).
		AddItem(rightBlock, 0, 1, false)

	u.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlC:
			close(u.monitor.GetQuit())
			u.app.Stop()
			return nil
		case tcell.KeyEscape:
			if u.monitor.IsShow() {
				u.monitor.SetShow(false)
				u.updateInfo()
			}
			return nil
		}
		return event
	})

	u.app.SetRoot(mainBlock, true).
		SetFocus(u.input)
}

func (u *UI) Run() error {
	u.setupUI()
	go u.start()
	return u.app.Run()
}

func (u *UI) start() {
	// * 每分鐘檢查一次
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	u.monitor.Update()
	u.app.QueueUpdateDraw(func() {
		u.updateUI()
	})

	for {
		select {
		case <-ticker.C:
			u.monitor.Update()
			u.app.QueueUpdateDraw(func() {
				u.updateUI()
			})
		case <-u.monitor.GetQuit():
			return
		}
	}
}

func (u *UI) updateUI() {
	u.table.Clear()

	headers := []string{"Website", "Status", "Duration", "Code", "Last Check", "SSL Expire"}
	for col, header := range headers {
		u.table.SetCell(0, col, tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter).
			SetSelectable(false).
			SetExpansion(1))
	}

	websites := u.monitor.GetList()
	statuses := u.monitor.GetStatus()

	row := 1
	for _, site := range websites {
		status, exists := statuses[site]
		if !exists {
			continue
		}

		u.table.SetCell(row, 0, tview.NewTableCell(status.URL).
			SetTextColor(tcell.ColorGreen).
			SetAlign(tview.AlignCenter))

		statusText := "offline"
		statusColor := tcell.ColorRed
		if status.Online {
			statusText = "online"
			statusColor = tcell.ColorGreen
		}
		u.table.SetCell(row, 1, tview.NewTableCell(statusText).
			SetTextColor(statusColor).
			SetAlign(tview.AlignCenter))

		u.table.SetCell(row, 2, tview.NewTableCell(fmt.Sprintf("%dms", status.Duration.Milliseconds())).
			SetAlign(tview.AlignCenter))

		codeText := "N/A"
		codeColor := tcell.ColorRed
		if status.Code > 0 {
			codeText = fmt.Sprintf("%d", status.Code)
			if status.Code >= 200 && status.Code < 400 {
				codeColor = tcell.ColorGreen
			}
		}
		u.table.SetCell(row, 3, tview.NewTableCell(codeText).
			SetTextColor(codeColor).
			SetAlign(tview.AlignCenter))

		u.table.SetCell(row, 4, tview.NewTableCell(status.LastCheck.Format("15:04:05")).
			SetAlign(tview.AlignCenter))

		sslText := "N/A"
		sslColor := tcell.ColorRed
		if status.Expire > 0 {
			sslText = fmt.Sprintf("%d Days", status.Expire)
			if status.Expire > 30 {
				sslColor = tcell.ColorGreen
			} else if status.Expire > 7 {
				sslColor = tcell.ColorYellow
			}
		}
		u.table.SetCell(row, 5, tview.NewTableCell(sslText).
			SetTextColor(sslColor).
			SetAlign(tview.AlignCenter))

		row++
	}

	u.updateInfo()
}

func (u *UI) updateInfo() {
	if u.monitor.IsShow() {
		u.showSMTP()
		return
	}

	list := u.monitor.GetStatus()
	config := u.monitor.GetConfig()

	total := len(list)
	online := 0
	for _, e := range list {
		if e.Online {
			online++
		}
	}

	uptime := float64(online) / float64(total) * 100
	if total == 0 {
		uptime = 0
	}

	status := "Disable"
	smtpColor := "red"
	if config.Enabled && config.Host != "" {
		status = "Enable"
		smtpColor = "green"
	}

	info := fmt.Sprintf(`[yellow]Website Monitoring System[-]
[white]Last Update:[-] %s

[white]Total Sites:[-] %d
[green]Online:[-] %d
[red]Offline:[-] %d
[cyan]Uptime:[-] %.1f%%
[white]Email Notification:[-] [%s]%s[-]

[gray]Commands:[-]
[white]add [-][yellow]<url>[-][gray] - Add website[-]
[white][delete|del|remove|rm] [-][yellow]<url>[-][gray] - Remove website[-]
[white]smtp[-][gray] - Show SMTP settings[-]
[white]smtp [-][yellow]<setting> <value>[-][gray] - Configure SMTP[-]
[white]test[-][gray] - Send test email[-]
[white]refresh[-][gray] - Manual update[-]
[white][quit|exit|bye][-][gray] - Exit program[-]`,
		time.Now().Format("2006-01-02 15:04:05"),
		total,
		online,
		total-online,
		uptime,
		smtpColor,
		status,
	)

	u.text.SetText(info)
}

func (u *UI) showSMTP() {
	config := u.monitor.GetConfig()

	password := "unset"
	if len(config.Password) > 0 {
		password = strings.Repeat("*", len(config.Password))
	}

	enable := "Disable"
	enableColor := "red"
	if config.Enabled {
		enable = "Enable"
		enableColor = "green"
	}

	toList := "unset"
	if len(config.To) > 0 {
		toList = strings.Join(config.To, ", ")
	}

	host := config.Host
	if host == "" {
		host = "unset"
	}

	username := config.Username
	if username == "" {
		username = "unset"
	}

	from := config.From
	if from == "" {
		from = "unset"
	}

	CC := config.CC
	if CC == "" {
		CC = "unset"
	}

	info := fmt.Sprintf(`[yellow]SMTP Email Config[-]

[white]Host:[-] %s
[white]Port:[-] %d
[white]Username:[-] %s
[white]Password:[-] %s
[white]From:[-] %s
[white]To:[-] %s
[white]CC:[-] %s
[white]Enabled:[-] [%s]%s[-]

[gray]Commands:[-]
[white]smtp host [-][yellow]<host>[-][gray] - Set host[-]
[white]smtp port [-][yellow]<port>[-][gray] - Set port[-]
[white]smtp username [-][yellow]<username>[-][gray] - Set username[-]
[white]smtp password [-][yellow]<password>[-][gray] - Set password[-]
[white]smtp from [-][yellow]<sender>[-][gray] - Set sender[-]
[white]smtp add [-][yellow]<email>[-][gray] - Add email[-]
[white]smtp [delete|del|remove|rm] [-][yellow]<email>[-][gray] - Remove email[-]
[white]smtp enabled [-][yellow][true|false|1|0][-][gray] - Enable/disable[-]
[white]test[-][gray] - Send test email[-]
[white]back[-][gray] - Return to main screen[-]
[white][quit|exit|bye][-][gray] - Exit program[-]`,
		host,
		config.Port,
		username,
		password,
		from,
		toList,
		CC,
		enableColor,
		enable,
	)

	u.text.SetText(info)
}

func (u *UI) handleCommand(cmd string) {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return
	}

	switch parts[0] {
	case "add":
		if len(parts) > 1 {
			url := parts[1]
			if err := u.monitor.Add(url); err != nil {
				fmt.Printf("%v\n", err)
				return
			}
			u.monitor.Save()
			u.updateUI()
		}

	case "delete", "del", "remove", "rm":
		if len(parts) > 1 {
			url := parts[1]
			if err := u.monitor.Remove(url); err != nil {
				fmt.Printf("%v\n", err)
				return
			}
			u.monitor.Save()
			u.updateUI()
		}

	case "smtp":
		if len(parts) < 2 {
			u.monitor.SetShow(true)
			u.updateInfo()
			return
		}

		if parts[1] == "show" {
			u.monitor.SetShow(true)
			u.updateInfo()
			return
		}

		if len(parts) < 3 {
			return
		}

		key := parts[1]
		value := parts[2]

		switch key {
		case "add":
			u.monitor.AddEmail(value)
		case "delete", "del", "remove", "rm":
			u.monitor.RemoveEmail(value)
		default:
			if err := u.monitor.UpdateSMTP(key, value); err != nil {
				fmt.Printf("%v\n", err)
				return
			}
		}

		u.monitor.Save()

		if u.monitor.IsShow() {
			u.updateInfo()
		}

	case "back":
		u.monitor.SetShow(false)
		u.updateInfo()

	case "test":
		go func() {
			if err := u.monitor.TestEmail(); err != nil {
				slog.Error("Failed to send test email", "error", err)
			}
		}()

	case "refresh":
		go func() {
			u.monitor.Update()
			u.app.QueueUpdateDraw(func() {
				u.updateUI()
			})
		}()

	case "quit", "exit", "bye":
		close(u.monitor.GetQuit())
		u.app.Stop()
	}
}
