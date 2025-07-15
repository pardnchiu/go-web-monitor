package service

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"website-monitor/internal/model"
	"website-monitor/internal/util"
)

type Monitor struct {
	list   []string
	status map[string]*model.Website
	config model.Config
	smtp   *SMTP
	mu     sync.RWMutex
	quit   chan bool
	isShow bool
}

func New(list []string, config model.Config) *Monitor {
	return &Monitor{
		list:   list,
		status: make(map[string]*model.Website),
		config: config,
		smtp:   NewSMTP(config),
		quit:   make(chan bool),
	}
}

func (m *Monitor) GetList() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.list
}

func (m *Monitor) GetStatus() map[string]*model.Website {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.status
}

func (m *Monitor) GetConfig() model.Config {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.config
}

func (m *Monitor) IsShow() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.isShow
}

func (m *Monitor) SetShow(isShow bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.isShow = isShow
}

func (m *Monitor) GetQuit() chan bool {
	return m.quit
}

func (m *Monitor) Check(url string) model.Website {
	start := time.Now()

	if len(url) > 4 && url[:4] != "http" {
		url = "https://" + url
	} else if len(url) <= 4 {
		url = "https://" + url
	}

	// * 超時 5 秒
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	res, err := client.Get(url)
	var expire int
	if err == nil {
		expire, _ = util.CheckExpire(url)
	}

	status := model.Website{
		URL:       url,
		Code:      0,
		Duration:  time.Since(start),
		Online:    false,
		LastCheck: time.Now(),
		Expire:    expire,
	}

	if err != nil {
		return status
	}
	defer res.Body.Close()

	status.Code = res.StatusCode
	status.Online = res.StatusCode >= 200 && res.StatusCode < 400

	return status
}

func (m *Monitor) Update() {
	var wg sync.WaitGroup

	for _, site := range m.list {
		wg.Add(1)

		go func(url string) {
			defer wg.Done()
			status := m.Check(url)
			prev := m.status[url]

			count := 0
			if prev != nil {
				m.mu.RLock()
				count = m.status[url].Count
				m.mu.RUnlock()
			}

			m.mu.Lock()
			// * 5 次失敗寄送 Email
			if !status.Online && (count+1)%5 == 0 {
				go m.sendEmail(url, fmt.Sprintf("HTTP %d", status.Code))
			}
			if status.Online || (count+1)%5 == 0 {
				m.status[url] = &model.Website{
					URL:       status.URL,
					Code:      status.Code,
					Duration:  status.Duration,
					Online:    status.Online,
					LastCheck: status.LastCheck,
					Expire:    status.Expire,
					Count:     0,
				}
			} else {
				m.status[url] = &model.Website{
					URL:       status.URL,
					Code:      status.Code,
					Duration:  status.Duration,
					Online:    status.Online,
					LastCheck: status.LastCheck,
					Expire:    status.Expire,
					Count:     count + 1,
				}
			}
			m.mu.Unlock()
		}(site)
	}

	wg.Wait()

	if err := m.Save(); err != nil {
		slog.Error("Failed to save monitor data", "error", err)
	}
}

func (m *Monitor) sendEmail(url, reason string) {
	subject := fmt.Sprintf("Website Offline Alert: %s", url)
	body := fmt.Sprintf(`<h1>Website Monitor</h1>
<hr>
<br>
Website: %s<br>
Status: Offline<br>
Reason: %s<br>
Check Time: %s<br>
<br>
Please check the website status immediately.<br>
<br>
<hr>
Monitoring service provided by <a href="https://github.com/pardnchiu/web-monitor">pardnchiu/web-monitor</a>
`, url, reason, time.Now().Format("2006-01-02 15:04:05"))
	if err := m.smtp.SendEmail(subject, body); err != nil {
		slog.Error("Failed to send email", "error", err)
	}
}

func (m *Monitor) GetWebList() []model.Website {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []model.Website
	for _, e := range m.status {
		list = append(list, *e)
	}
	return list
}

func (m *Monitor) Save() error {
	monitorData := model.MonitorData{
		List:   m.GetWebList(),
		Config: m.GetConfig(),
	}
	return util.Save(".webMonitor.json", monitorData)
}

func (m *Monitor) Add(url string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, e := range m.list {
		if e == url {
			return fmt.Errorf("already exists: %s", url)
		}
	}

	status := m.Check(url)
	m.status[url] = &status
	m.list = append(m.list, url)
	return nil
}

func (m *Monitor) Remove(url string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	isExist := false
	for _, e := range m.list {
		if e == url {
			isExist = true
			break
		}
	}

	if !isExist {
		return fmt.Errorf("not found: %s", url)
	}

	delete(m.status, url)
	for i, site := range m.list {
		if site == url {
			m.list = append(m.list[:i], m.list[i+1:]...)
			break
		}
	}
	return nil
}

func (m *Monitor) UpdateSMTP(key, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch key {
	case "host":
		m.config.Host = value
	case "port":
		port, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid port: %s", value)
		}
		m.config.Port = port
	case "username":
		m.config.Username = value
	case "password":
		m.config.Password = value
	case "from":
		m.config.From = value
	case "enabled":
		m.config.Enabled = value == "true" || value == "1"
	default:
		return fmt.Errorf("unknown key: %s", key)
	}

	m.smtp.UpdateConfig(m.config)
	return nil
}

func (m *Monitor) AddEmail(email string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	email = strings.TrimSpace(email)
	if email == "" {
		return
	}

	for _, e := range m.config.To {
		if e == email {
			return
		}
	}

	m.config.To = append(m.config.To, email)
	m.smtp.UpdateConfig(m.config)
}

func (m *Monitor) RemoveEmail(email string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	email = strings.TrimSpace(email)
	if email == "" {
		return
	}

	newTo := []string{}
	for _, e := range m.config.To {
		if e != email {
			newTo = append(newTo, e)
		}
	}
	m.config.To = newTo
	m.smtp.UpdateConfig(m.config)
}

func (m *Monitor) TestEmail() error {
	subject := "Website Monitoring System - Test"
	body := fmt.Sprintf(`<h1>Website Monitor</h1>
<hr>
<br>
This is a test email from the website monitoring system.<br>
<br>
If you receive this email, it means the SMTP configuration is correct.<br>
<br>
Test time: %s<br>
<br>
<hr>
Monitoring service provided by <a href="https://github.com/pardnchiu/web-monitor">pardnchiu/web-monitor</a>
`, time.Now().Format("2006-01-02 15:04:05"))

	return m.smtp.SendEmail(subject, body)
}
