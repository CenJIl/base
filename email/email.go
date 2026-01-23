package email

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"strings"
)

type QQMail struct {
	From     string
	Password string
}

func NewQQMail(from, password string) *QQMail {
	return &QQMail{
		From:     from,
		Password: password,
	}
}

func (m *QQMail) Send(to []string, subject, body string) error {
	smtpHost := "smtp.qq.com"
	smtpPort := "465"

	// 1. 先建立TLS连接（465端口是SSL直连，不是STARTTLS）
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         smtpHost,
	}
	conn, err := tls.Dial("tcp", smtpHost+":"+smtpPort, tlsConfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	// 2. 用TLS连接创建SMTP客户端
	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		return err
	}
	defer client.Close()

	// 3. 认证
	auth := smtp.PlainAuth("", m.From, m.Password, smtpHost)
	if err = client.Auth(auth); err != nil {
		return err
	}

	// 4. 发件人、收件人
	if err = client.Mail(m.From); err != nil {
		return err
	}
	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}

	// 5. 写入邮件（加MIME头防中文乱码）
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
