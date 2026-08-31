package email

import (
	"strings"
	"testing"
)

func TestSendRejectsEmptyRecipientsBeforeDialing(t *testing.T) {
	mail := NewQQMail("sender@qq.com", "authorization-code")
	err := mail.Send(nil, "subject", "body")
	if err == nil || !strings.Contains(err.Error(), "收件人列表不能为空") {
		t.Fatalf("Send() error = %v, want empty recipient error", err)
	}
}
