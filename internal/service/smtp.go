package service

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"website-monitor/internal/model"
)

type SMTP struct {
	config model.Config
}

func NewSMTP(cfg model.Config) *SMTP {
	return &SMTP{
		config: cfg,
	}
}

func (c *SMTP) UpdateConfig(cfg model.Config) {
	c.config = cfg
}

func (c *SMTP) SendEmail(subject, body string) error {
	if !c.config.Enabled || c.config.Host == "" {
		return fmt.Errorf("SMTP not Enabled or Host is not set")
	}

	if len(c.config.To) == 0 {
		return fmt.Errorf("no user email configured")
	}

	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)

	tcpConn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect TCP (%s): %v", addr, err)
	}
	tcpConn.Close()

	var conn *smtp.Client

	conn, err = smtp.Dial(addr)
	if err != nil {
		if c.config.Port == 465 {
			tlsConfig := &tls.Config{
				ServerName: c.config.Host,
			}

			tlsConn, err := tls.Dial("tcp", addr, tlsConfig)
			if err != nil {
				return fmt.Errorf("failed to connect TLS (%s): %v", addr, err)
			}

			conn, err = smtp.NewClient(tlsConn, c.config.Host)
			if err != nil {
				tlsConn.Close()
				return fmt.Errorf("failed to create SMTP client: %v", err)
			}
		} else {
			return fmt.Errorf("failed to connect SMTP (%s): %v", addr, err)
		}
	}
	defer conn.Close()

	// STARTTLS for port 587
	if c.config.Port == 587 {
		if ok, _ := conn.Extension("STARTTLS"); ok {
			if err = conn.StartTLS(&tls.Config{
				ServerName: c.config.Host,
			}); err != nil {
				return fmt.Errorf("failed to start TLS: %v", err)
			}
		}
	}

	if c.config.Username != "" && c.config.Password != "" {
		auth := smtp.PlainAuth("", c.config.Username, c.config.Password, c.config.Host)
		if err = conn.Auth(auth); err != nil {
			return fmt.Errorf("failed to authenticate: %v", err)
		}
	}

	if err = conn.Mail(c.config.From); err != nil {
		return fmt.Errorf("failed to set sender (%s): %v", c.config.From, err)
	}

	for _, to := range c.config.To {
		if err = conn.Rcpt(to); err != nil {
			return fmt.Errorf("failed to set email (%s): %v", to, err)
		}
	}

	wc, err := conn.Data()
	if err != nil {
		return fmt.Errorf("failed to get writer: %v", err)
	}
	defer wc.Close()

	if c.config.CC == "" {
		c.config.CC = c.config.From
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nCc: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		c.config.From,
		strings.Join(c.config.To, ","),
		c.config.CC,
		// strings.Join(c.config.CC, ","),
		subject,
		body)

	if _, err = wc.Write([]byte(msg)); err != nil {
		return fmt.Errorf("failed to write email: %v", err)
	}

	return nil
}
