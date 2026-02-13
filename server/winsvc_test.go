//go:build windows

package server

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/CenJIl/base/common"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/windows/svc"
)

func TestDefault(t *testing.T) {
	svc := DefaultWinSVC(GetName)
	assert.NotNil(t, svc)
	assert.Contains(t, svc.Name, "GetName")
	fmt.Printf("%v", svc)
}

func TestDefaultWinSVC_HandlerNameExtraction(t *testing.T) {
	svc := DefaultWinSVC(SimpleHandler)
	assert.Equal(t, "SimpleHandler", svc.Name)
	assert.Equal(t, "SimpleHandler", svc.DisplayName)
	assert.Contains(t, svc.Description, "SimpleHandler")
	assert.Equal(t, 15*time.Second, svc.ShutdownWait)
}

func TestDefaultWinSVC_NestedName(t *testing.T) {
	svc := DefaultWinSVC(testHandler)
	assert.Equal(t, "testHandler", svc.Name)
}

func testHandler(ctx context.Context) error {
	return nil
}

func TestDefaultWinSVC_HandlerWithPackage(t *testing.T) {
	s := &serverType{}
	svc := DefaultWinSVC(s.PackageHandler)
	assert.Contains(t, svc.Name, "PackageHandler")
}

type serverType struct{}

func (s *serverType) PackageHandler(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func TestWinSVC_String(t *testing.T) {
	svc := &WinSVC{
		Name:        "TestService",
		DisplayName: "测试服务",
		Description: "测试描述",
	}

	str := svc.String()
	assert.Contains(t, str, "TestService")
	assert.Contains(t, str, "测试服务")
	assert.Contains(t, str, "测试描述")
}

func TestWinSVC_String_DefaultValues(t *testing.T) {
	svc := DefaultWinSVC(SimpleHandler)
	svc.Name = "MyService"
	svc.DisplayName = "My Service"
	svc.Description = "My Description"

	str := svc.String()
	assert.Contains(t, str, "MyService")
	assert.Contains(t, str, "My Service")
	assert.Contains(t, str, "My Description")
}

func TestWinSVC_Execute_Simple(t *testing.T) {
	handlerDone := make(chan struct{})

	handler := func(ctx context.Context) error {
		<-ctx.Done()
		close(handlerDone)
		return nil
	}

	winSvc := &WinSVC{
		Name:         "TestSimpleService",
		DisplayName:  "Test Simple Service",
		Handler:      handler,
		ShutdownWait: 1 * time.Second,
		Log:          &common.DefaultLog{},
	}

	stopReq := make(chan svc.ChangeRequest, 1)
	statusChan := make(chan svc.Status, 10)

	go func() {
		time.Sleep(100 * time.Millisecond)
		stopReq <- svc.ChangeRequest{Cmd: svc.Stop}
	}()

	cont, ec := winSvc.Execute([]string{}, stopReq, statusChan)
	assert.False(t, cont)
	assert.Equal(t, uint32(0), ec)

	select {
	case <-handlerDone:
	case <-time.After(2 * time.Second):
		t.Fatal("Handler did not complete")
	}
}

func TestWinSVC_CustomFields(t *testing.T) {
	customHandler := func(ctx context.Context) error {
		return nil
	}

	svc := &WinSVC{
		Name:         "CustomService",
		DisplayName:  "自定义服务",
		Description:  "自定义描述",
		ShutdownWait: 30 * time.Second,
		Handler:      customHandler,
	}

	assert.Equal(t, "CustomService", svc.Name)
	assert.Equal(t, "自定义服务", svc.DisplayName)
	assert.Equal(t, "自定义描述", svc.Description)
	assert.Equal(t, 30*time.Second, svc.ShutdownWait)
}

func SimpleHandler(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func GetName(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		fmt.Println(" ")
	}
}
