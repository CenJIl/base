package email

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testEmailFrom     = "1415738397@qq.com"
	testEmailPassword = "ehhyswajbeiybaea"
	testEmailTo       = "476503904@qq.com"
)

func TestNewQQMail(t *testing.T) {
	mail := NewQQMail(testEmailFrom, testEmailPassword)

	assert.NotNil(t, mail)
	assert.Equal(t, testEmailFrom, mail.From)
	assert.Equal(t, testEmailPassword, mail.Password)
	assert.Equal(t, 10*time.Second, mail.Timeout)
}

func TestNewQQMail_DefaultTimeout(t *testing.T) {
	mail := NewQQMail(testEmailFrom, testEmailPassword)

	assert.Equal(t, 10*time.Second, mail.Timeout)
}

func TestSendEmail_RealEmail(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real email test in short mode")
	}

	mail := NewQQMail(testEmailFrom, testEmailPassword)
	err := mail.Send([]string{testEmailTo}, "测试邮件-"+time.Now().Format("20060102150405"), "这是一封测试邮件，用于测试邮件发送功能。")

	require.NoError(t, err)
}

func TestSendEmail_ChineseSubject(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real email test in short mode")
	}

	mail := NewQQMail(testEmailFrom, testEmailPassword)
	chineseSubject := "中文主题测试 - 包含特殊字符：！@#￥%……&*（）"

	err := mail.Send([]string{testEmailTo}, chineseSubject, "邮件正文内容")

	require.NoError(t, err)
}

func TestSendEmail_ChineseBody(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real email test in short mode")
	}

	mail := NewQQMail(testEmailFrom, testEmailPassword)
	chineseBody := "中文邮件正文测试\n第二行内容\n第三行内容：包含数字123和英文abc"

	err := mail.Send([]string{testEmailTo}, "中文正文测试", chineseBody)

	require.NoError(t, err)
}

func TestSendEmail_EmptySubject(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real email test in short mode")
	}

	mail := NewQQMail(testEmailFrom, testEmailPassword)
	err := mail.Send([]string{testEmailTo}, "", "邮件内容为空主题")

	require.NoError(t, err)
}

func TestSendEmail_EmptyBody(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real email test in short mode")
	}

	mail := NewQQMail(testEmailFrom, testEmailPassword)
	err := mail.Send([]string{testEmailTo}, "空正文测试", "")

	require.NoError(t, err)
}

func TestSendEmail_MultipleRecipients(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real email test in short mode")
	}

	mail := NewQQMail(testEmailFrom, testEmailPassword)
	recipients := []string{testEmailTo, testEmailFrom}
	subject := "多收件人测试 - " + time.Now().Format("20060102150405")

	err := mail.Send(recipients, subject, "这是一封发送给多个收件人的测试邮件")

	require.NoError(t, err)
}

func TestSendEmail_LongSubject(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real email test in short mode")
	}

	mail := NewQQMail(testEmailFrom, testEmailPassword)
	longSubject := strings.Repeat("这是一个很长的主题-", 50)

	err := mail.Send([]string{testEmailTo}, longSubject, "长主题测试")

	require.NoError(t, err)
}

func TestSendEmail_LongBody(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real email test in short mode")
	}

	mail := NewQQMail(testEmailFrom, testEmailPassword)
	longBody := strings.Repeat("这是一行很长的邮件内容。", 100)

	err := mail.Send([]string{testEmailTo}, "长正文测试", longBody)

	require.NoError(t, err)
}

func TestSendEmail_SpecialCharacters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real email test in short mode")
	}

	mail := NewQQMail(testEmailFrom, testEmailPassword)
	specialSubject := "特殊字符测试 !@#$%^&*()_+-=[]{}|;':\",./<>?"
	specialBody := "特殊字符内容\n\t测试制表符\n测试换行符"

	err := mail.Send([]string{testEmailTo}, specialSubject, specialBody)

	require.NoError(t, err)
}

func TestSendEmail_Unicode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real email test in short mode")
	}

	mail := NewQQMail(testEmailFrom, testEmailPassword)
	unicodeSubject := "Unicode测试 🎉🚀⭐✨"
	unicodeBody := "Emoji测试 😀😊😎\n符号测试 ✓✗★♥\n组合测试 🌟🎊🎈"

	err := mail.Send([]string{testEmailTo}, unicodeSubject, unicodeBody)

	require.NoError(t, err)
}

func TestSendEmail_MultiLineBody(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real email test in short mode")
	}

	mail := NewQQMail(testEmailFrom, testEmailPassword)
	multiLineBody := `第一行内容
第二行内容
第三行内容
第四行内容
第五行内容`

	err := mail.Send([]string{testEmailTo}, "多行正文测试", multiLineBody)

	require.NoError(t, err)
}

func TestSendEmail_CustomTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real email test in short mode")
	}

	mail := NewQQMail(testEmailFrom, testEmailPassword)
	mail.Timeout = 15 * time.Second

	err := mail.Send([]string{testEmailTo}, "自定义超时测试", "测试15秒超时设置")

	require.NoError(t, err)
}

func TestSendEmail_VeryShortTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real email test in short mode")
	}

	mail := NewQQMail(testEmailFrom, testEmailPassword)
	mail.Timeout = 5 * time.Second

	err := mail.Send([]string{testEmailTo}, "短超时测试", "测试5秒超时设置")

	require.NoError(t, err)
}

func TestSendEmail_HtmlLikeContent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real email test in short mode")
	}

	mail := NewQQMail(testEmailFrom, testEmailPassword)
	htmlLikeBody := `包含HTML标签的文本（虽然当前不支持HTML，但测试文本内容）：
<div>这是div标签</div>
<p>这是p标签</p>
<span>这是span标签</span>`

	err := mail.Send([]string{testEmailTo}, "HTML内容测试", htmlLikeBody)

	require.NoError(t, err)
}

func TestSendEmail_NewlineFormats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real email test in short mode")
	}

	mail := NewQQMail(testEmailFrom, testEmailPassword)
	bodyWithNewlines := "Unix换行符\n\n\r\nWindows换行符\n\n回车符\r\r"

	err := mail.Send([]string{testEmailTo}, "换行符测试", bodyWithNewlines)

	require.NoError(t, err)
}

func TestSendEmail_TimestampInSubject(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real email test in short mode")
	}

	mail := NewQQMail(testEmailFrom, testEmailPassword)
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	subject := "时间戳测试 - " + timestamp

	err := mail.Send([]string{testEmailTo}, subject, "测试邮件主题中包含时间戳")

	require.NoError(t, err)
}

func TestSendEmail_ConsecutiveEmails(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real email test in short mode")
	}

	mail := NewQQMail(testEmailFrom, testEmailPassword)

	for i := 0; i < 3; i++ {
		subject := "连续邮件测试 " + time.Now().Format("20060102150405")
		err := mail.Send([]string{testEmailTo}, subject, "这是第"+string(rune('1'+i))+"封连续测试邮件")
		require.NoError(t, err)

		time.Sleep(1 * time.Second)
	}
}

func TestSendEmail_StructValidation(t *testing.T) {
	mail := NewQQMail(testEmailFrom, testEmailPassword)

	assert.NotEmpty(t, mail.From)
	assert.NotEmpty(t, mail.Password)
	assert.Greater(t, mail.Timeout, 0*time.Second)
}

func BenchmarkNewQQMail(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewQQMail(testEmailFrom, testEmailPassword)
	}
}
