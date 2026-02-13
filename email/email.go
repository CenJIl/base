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

// QQMail QQ 邮箱 SMTP 客户端
//
// 提供使用 QQ 邮箱 SMTP 服务发送邮件的功能
// 支持 SSL/TLS 加密连接和超时控制
type QQMail struct {
	From     string        // 发件人邮箱地址（QQ 邮箱）
	Password string        // 邮箱密码或授权码（需要在 QQ 邮箱设置中开启 SMTP 服务）
	Timeout  time.Duration // 连接和握手超时时间，默认 10 秒
}

// NewQQMail 创建 QQ 邮件客户端
//
// 使用指定的发件人邮箱和密码创建一个新的邮件客户端实例
// 连接超时时间默认设置为 10 秒
//
// 参数
//
//	from - 发件人 QQ 邮箱地址（例如：123456789@qq.com）
//	password - QQ 邮箱密码或授权码（需在 QQ 邮箱设置中生成授权码）
//
// 返回值
//
//	*QQMail - 邮件客户端实例
//
// 注意事项
//   - 密码建议使用 QQ 邮箱的授权码而非真实密码
//   - 需要提前在 QQ 邮箱设置中开启 POP3/SMTP 服务并生成授权码
//   - 默认连接超时为 10 秒，可根据需要修改 Timeout 字段
//   - QQ 邮箱 SMTP 服务器：smtp.qq.com，端口：465
//
// 示例
//
//	mail := email.NewQQMail("your@qq.com", "your-auth-code")
//	mail.Timeout = 15 * time.Second  // 可选：修改超时时间
func NewQQMail(from, password string) *QQMail {
	return &QQMail{
		From:     from,
		Password: password,
		Timeout:  10 * time.Second,
	}
}

// Send 发送邮件
//
// 通过 QQ 邮箱 SMTP 服务发送邮件，支持中文内容（使用 Base64 编码）
// 使用 SSL/TLS 加密连接确保传输安全
//
// 参数
//
//	to - 收件人邮箱地址列表（支持多个收件人）
//	subject - 邮件主题（支持中文，会自动进行 Base64 编码）
//	body - 邮件正文内容（纯文本格式，支持中文）
//
// 返回值
//
//	error - 发送失败时返回错误信息，成功返回 nil
//
// 错误类型
//   - 连接超时或失败
//   - TLS 握手失败
//   - SMTP 认证失败
//   - 邮件发送失败
//
// 注意事项
//   - 收件人列表不能为空
//   - 邮件主题和正文支持中文，会自动处理编码
//   - 连接和 TLS 握手都有超时控制（Timeout 字段）
//   - 每次发送都会建立新的 SSL/TLS 连接
//   - QQ 邮箱有发送频率限制，频繁发送可能被限制
//
// 示例
//
//	mail := email.NewQQMail("from@qq.com", "password")
//	err := mail.Send(
//	    []string{"to1@qq.com", "to2@qq.com"},
//	    "测试邮件",
//	    "这是一封测试邮件内容",
//	)
//	if err != nil {
//	    log.Fatalf("邮件发送失败: %v", err)
//	}
func (m *QQMail) Send(to []string, subject, body string) error {
	smtpHost := "smtp.qq.com"
	smtpPort := "465"

	conn, err := net.DialTimeout("tcp", smtpHost+":"+smtpPort, m.Timeout)
	if err != nil {
		return fmt.Errorf("连接超时或失败: %w", err)
	}
	defer conn.Close()

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         smtpHost,
	}
	tlsConn := tls.Client(conn, tlsConfig)

	if err := tlsConn.SetDeadline(time.Now().Add(m.Timeout)); err != nil {
		return fmt.Errorf("设置超时失败: %w", err)
	}
	if err := tlsConn.Handshake(); err != nil {
		return fmt.Errorf("TLS 握手失败: %w", err)
	}
	tlsConn.SetDeadline(time.Time{})

	client, err := smtp.NewClient(tlsConn, smtpHost)
	if err != nil {
		return err
	}
	defer client.Close()

	auth := smtp.PlainAuth("", m.From, m.Password, smtpHost)
	if err = client.Auth(auth); err != nil {
		return err
	}

	if err = client.Mail(m.From); err != nil {
		return err
	}
	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}

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
