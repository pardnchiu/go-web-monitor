package model

import (
	"time"
)

type Website struct {
	URL       string        `json:"url"`
	Code      int           `json:"code"`
	Duration  time.Duration `json:"duration"`
	Online    bool          `json:"online"`
	LastCheck time.Time     `json:"last_check"`
	Expire    int           `json:"expire"`
	Count     int           `json:"count"`
}

type Config struct {
	Host     string   `json:"host"`
	Port     int      `json:"port"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	From     string   `json:"from"`
	To       []string `json:"to"`
	// CC       []string `json:"cc"`
	CC      string `json:"cc"`
	Enabled bool   `json:"enabled"`
}

type MonitorData struct {
	List   []Website `json:"list"`
	Config Config    `json:"config"`
}
