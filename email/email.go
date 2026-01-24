package email

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type QQMail struct {
	From     string
	Password string
	Timeout  time.Duration // 连接超时时间
}

func NewQQMail(from, password string) *QQMail {
	return &QQMail{
		From:     from,
		Password: password,
		Timeout:  10 * time.Second, // 默认 10 秒超时
	}
}

func (m *QQMail) Send(to []string, subject, body string) error {
	smtpHost := "smtp.qq.com"
	smtpPort := "465"

	// 1. 先建立 TCP 连接（带超时）
	conn, err := net.DialTimeout("tcp", smtpHost+":"+smtpPort, m.Timeout)
	if err != nil {
		return fmt.Errorf("连接超时或失败: %w", err)
	}
	defer conn.Close()

	// 2. 升级为 TLS 连接
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         smtpHost,
	}
	tlsConn := tls.Client(conn, tlsConfig)

	// 3. TLS 握手（带超时）
	if err := tlsConn.SetDeadline(time.Now().Add(m.Timeout)); err != nil {
		return fmt.Errorf("设置超时失败: %w", err)
	}
	if err := tlsConn.Handshake(); err != nil {
		return fmt.Errorf("TLS 握手失败: %w", err)
	}
	// 握手完成后清除超时，避免后续操作超时
	tlsConn.SetDeadline(time.Time{})

	// 4. 用 TLS 连接创建 SMTP 客户端
	client, err := smtp.NewClient(tlsConn, smtpHost)
	if err != nil {
		return err
	}
	defer client.Close()

	// 5. 认证
	auth := smtp.PlainAuth("", m.From, m.Password, smtpHost)
	if err = client.Auth(auth); err != nil {
		return err
	}

	// 6. 发件人、收件人
	if err = client.Mail(m.From); err != nil {
		return err
	}
	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}

	// 7. 写入邮件（加MIME头防中文乱码）
	w, err := client.Data()
	if err != nil {
		return err
	}

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: =?UTF-8?B?%s?=\r\n"+
			"MIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		m.From,
		strings.Join(to, ","),
		base64.StdEncoding.EncodeToString([]byte(subject)),
		body,
	)

	_, err = w.Write([]byte(msg))
	w.Close()
	return err
}
