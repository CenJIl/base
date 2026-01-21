//go:build windows

package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

type SVCLogger interface {
	Infof(format string, v ...any)
	Errorf(format string, v ...any)
}

type defaultLog struct{}

func (l *defaultLog) Infof(f string, v ...any)  { log.Printf("INFO  "+f, v...) }
func (l *defaultLog) Errorf(f string, v ...any) { log.Printf("ERROR "+f, v...) }

// WinSVC 核心结构体
type WinSVC struct {
	Name         string
	DisplayName  string
	Description  string
	SVCLog       SVCLogger
	ShutdownWait time.Duration
	Handler      func(ctx context.Context) error
}

func (w *WinSVC) String() string {
	return fmt.Sprintf(`{"Name":"%s","DisplayName":"%s","Description":"%s"}`, w.Name, w.DisplayName, w.Description)
}

// 创建默认 Windows 服务实例，使用15s退出等待时间。
// 需要自定义 DisplayName、Description、ShutdownWait 等参数时，直接构造 [WinSVC] 结构体
func DefaultWinSVC(handler func(ctx context.Context) error) *WinSVC {
	pc := reflect.ValueOf(handler).Pointer()
	fn := runtime.FuncForPC(pc)
	defaultName := "Default Service"
	if fn != nil {
		defaultName = fn.Name()
	}
	if idx := strings.LastIndex(defaultName, "."); idx != -1 {
		defaultName = defaultName[idx+1:]
	}
	return &WinSVC{
		Name:         defaultName,
		DisplayName:  defaultName,
		Description:  fmt.Sprintf("%s Create With Default", defaultName),
		SVCLog:       &defaultLog{},
		ShutdownWait: 15 * time.Second,
		Handler:      handler,
	}
}

// Run 启动服务（阻塞运行）
func (w *WinSVC) Run() {
	if w.SVCLog == nil {
		w.SVCLog = &defaultLog{}
	}

	ensureWorkingDirectory()

	if err := svc.Run(w.Name, w); err != nil {
		w.SVCLog.Errorf("服务 [%s] 启动失败: %v", w.Name, err)
		if el, err := eventlog.Open(w.Name); err == nil {
			_ = el.Error(1, fmt.Sprintf("Service failed: %v", err))
			el.Close()
		}
	}
}

// Execute 实现了 svc.Handler 接口
func (w *WinSVC) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	const cmds = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- w.Handler(ctx)
	}()

	changes <- svc.Status{State: svc.Running, Accepts: cmds}
	w.SVCLog.Infof("服务 [%s] 运行中...", w.Name)

	for {
		select {
		case err := <-errChan:
			if err != nil {
				w.SVCLog.Errorf("业务执行报错: %v", err)
				return false, 1
			}
			return false, 0
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				w.SVCLog.Infof("收到停止信号，执行优雅退出")
				cancel()
				changes <- svc.Status{
					State:    svc.StopPending,
					WaitHint: uint32(w.ShutdownWait.Milliseconds()),
				}
				select {
				case <-errChan:
				case <-time.After(w.ShutdownWait):
					w.SVCLog.Errorf("优雅退出超时")
				}
				return false, 0
			}
		}
	}
}

// Install 安装服务
//
//	// 用例: mysvc := DefaultWinSVC(myHandler); mysvc.Install()
//
//	// 使用示例:
//	//	flag := os.Args[1]
//	//	switch flag {
//	//	case "run":
//	//		svc.Run()
//	//	case "install":
//	//		svc.Install()
//	//	}
func (w *WinSVC) Install() {
	if !checkAndElevate() {
		return
	}

	m, err := mgr.Connect()
	if err != nil {
		log.Fatalf("连接管理器失败: %v", err)
	}
	defer m.Disconnect()

	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	s, err := m.CreateService(w.Name, exePath, mgr.Config{
		DisplayName: w.DisplayName,
		Description: w.Description,
		StartType:   mgr.StartAutomatic,
	}, "run")
	if err != nil {
		log.Fatalf("创建服务失败: %v", err)
	}
	defer s.Close()

	recovery := []mgr.RecoveryAction{
		{Type: mgr.ServiceRestart, Delay: time.Minute},
		{Type: mgr.ServiceRestart, Delay: time.Minute},
	}
	_ = s.SetRecoveryActions(recovery, 86400)

	log.Printf("服务 [%s] 安装成功", w.Name)
}

// Remove 卸载服务
func (w *WinSVC) Remove() {
	if !checkAndElevate() {
		return
	}

	m, err := mgr.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(w.Name)
	if err != nil {
		log.Fatalf("服务不存在: %v", err)
	}
	defer s.Close()

	if err := s.Delete(); err != nil {
		log.Fatalf("卸载失败: %v", err)
	}
	log.Printf("服务 [%s] 已卸载", w.Name)
}

// 辅助函数（提权、路径切换等保持不变）
func ensureWorkingDirectory() {
	exe, _ := os.Executable()
	_ = os.Chdir(filepath.Dir(exe))
}

func checkAndElevate() bool {
	if isAdmin() {
		return true
	}
	rerunAsAdmin()
	return false
}

func isAdmin() bool {
	var sid *windows.SID
	_ = windows.AllocateAndInitializeSid(&windows.SECURITY_NT_AUTHORITY, 2, windows.SECURITY_BUILTIN_DOMAIN_RID, windows.DOMAIN_ALIAS_RID_ADMINS, 0, 0, 0, 0, 0, 0, &sid)
	defer windows.FreeSid(sid)
	token := windows.Token(0)
	member, _ := token.IsMember(sid)
	return member
}

func rerunAsAdmin() {
	verb, _ := windows.UTF16PtrFromString("runas")
	exe, _ := os.Executable()
	exePtr, _ := windows.UTF16PtrFromString(exe)
	cwd, _ := os.Getwd()
	cwdPtr, _ := windows.UTF16PtrFromString(cwd)
	argPtr, _ := windows.UTF16PtrFromString(strings.Join(os.Args[1:], " "))
	_ = windows.ShellExecute(0, verb, exePtr, argPtr, cwdPtr, windows.SW_NORMAL)
	os.Exit(0)
}
