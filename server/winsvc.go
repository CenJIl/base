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

	"github.com/CenJIl/base/common"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

// WinSVC Windows 服务核心结构体
//
// 封装了 Windows 服务相关的所有属性和行为
// 提供服务的创建、安装、运行和卸载功能
type WinSVC struct {
	Name         string                          // 服务名称（系统唯一标识）
	DisplayName  string                          // 服务显示名称（服务管理器中显示）
	Description  string                          // 服务描述信息
	Log          common.Logger                   // 日志记录器，用于记录服务运行日志
	ShutdownWait time.Duration                   // 优雅关闭等待时间，默认 15 秒
	Handler      func(ctx context.Context) error // 服务主处理函数，在服务启动时执行
}

// String 返回服务的 JSON 格式字符串表示
//
// 包含服务的 Name、DisplayName 和 Description 字段
// 用于日志输出和调试
//
// 返回值
//
//	string - JSON 格式的服务信息字符串
//
// 示例
//
//	svc := &WinSVC{Name: "MyService", DisplayName: "My Windows Service"}
//	fmt.Println(svc.String())
//	// 输出: {"Name":"MyService","DisplayName":"My Windows Service","Description":""}
func (w *WinSVC) String() string {
	return fmt.Sprintf(`{"Name":"%s","DisplayName":"%s","Description":"%s"}`, w.Name, w.DisplayName, w.Description)
}

// DefaultWinSVC 使用默认参数创建 Windows 服务实例
//
// 使用传入的 handler 函数自动提取服务名称（基于函数名）
// 提供合理的默认值，适合大多数 Windows 服务场景
//
// 参数
//
//	handler - 服务主处理函数，接收 context 用于优雅关闭
//
// 返回值
//
//	*WinSVC - Windows 服务实例，使用默认配置
//
// 默认配置
//   - Name: 从 handler 函数名提取（例如 "MyHandler" -> "MyHandler"）
//   - DisplayName: 与 Name 相同
//   - Description: "{Name} Create With Default"
//   - ShutdownWait: 15 秒
//   - Log: common.DefaultLog（使用标准库 log）
//
// 注意事项
//   - handler 函数名会被作为服务名称，建议使用有意义的函数名
//   - 函数名包含包名时会自动去除包名部分
//   - handler 应该监听 ctx.Done() 来实现优雅关闭
//   - handler 返回错误会导致服务退出码为 1
//   - 默认日志记录器使用标准库 log，可自定义
//
// 示例
//
//	svc := server.DefaultWinSVC(myHandler)
//	svc.DisplayName = "我的服务"  // 可选：自定义显示名称
//	svc.Description = "这是一个示例服务"  // 可选：添加描述
//	svc.Run()
//
//	func myHandler(ctx context.Context) error {
//	    for {
//	        select {
//	        case <-ctx.Done():
//	            return nil
//	        default:
//	            // 执行服务逻辑
//	        }
//	    }
//	}
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
		Log:          &common.DefaultLog{},
		ShutdownWait: 15 * time.Second,
		Handler:      handler,
	}
}

// Run 启动服务（阻塞运行）
//
// 将当前程序注册为 Windows 服务并启动
// 此方法是阻塞的，直到服务停止才会返回
//
// 注意事项
//   - 此方法会阻塞，应该在 main 函数的最后一行调用
//   - 必须以管理员身份运行才能启动服务
//   - 如果 Log 为 nil，会自动使用 common.DefaultLog
//   - 启动失败会记录到 Windows 事件日志
//   - 服务启动前会自动切换工作目录到可执行文件所在目录
//
// 错误处理
//   - 服务启动失败会记录错误并退出
//   - 尝试写入 Windows 事件日志（如果可用）
//
// 示例
//
//	func main() {
//	    svc := server.DefaultWinSVC(myHandler)
//	    svc.Run()
//	}
func (w *WinSVC) Run() {
	if w.Log == nil {
		w.Log = &common.DefaultLog{}
	}

	ensureWorkingDirectory()

	if err := svc.Run(w.Name, w); err != nil {
		w.Log.Errorf("服务 [%s] 启动失败: %v", w.Name, err)
		if el, err := eventlog.Open(w.Name); err == nil {
			_ = el.Error(1, fmt.Sprintf("Service failed: %v", err))
			el.Close()
		}
	}
}

// Execute 实现 svc.Handler 接口，由 Windows 服务管理器调用
//
// 此方法是 Windows 服务生命周期的主要入口点，处理服务的启动、运行和停止
// 不需要手动调用，由服务管理器在适当时候调用
//
// 参数
//
//	args - 启动参数（通常为空）
//	r - 服务控制请求通道（停止、暂停等）
//	changes - 服务状态更新通道
//
// 返回值
//
//	bool - 是否继续运行
//	uint32 - 服务退出码（0 表示成功）
//
// 服务状态流转
//
//	StartPending -> Running -> StopPending -> Stopped
//
// 注意事项
//   - handler 函数在独立 goroutine 中运行
//   - 收到停止信号时会取消 context，handler 应监听 ctx.Done()
//   - 如果 handler 返回错误，退出码为 1
//   - 优雅关闭超时后会强制退出
//   - 支持 Interrogate、Stop、Shutdown 命令
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
	w.Log.Infof("服务 [%s] 运行中...", w.Name)

	for {
		select {
		case err := <-errChan:
			if err != nil {
				w.Log.Errorf("业务执行报错: %v", err)
				return false, 1
			}
			return false, 0
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				w.Log.Infof("收到停止信号，执行优雅退出")
				cancel()
				changes <- svc.Status{
					State:    svc.StopPending,
					WaitHint: uint32(w.ShutdownWait.Milliseconds()),
				}
				select {
				case <-errChan:
				case <-time.After(w.ShutdownWait):
					w.Log.Errorf("优雅退出超时")
				}
				return false, 0
			}
		}
	}
}

// Install 安装 Windows 服务
//
// 将当前可执行文件注册为 Windows 服务
// 服务配置为自动启动
//
// 注意事项
//   - 必须以管理员身份运行
//   - 如果不是管理员，会自动尝试提升权限
//   - 服务配置为自动启动类型
//   - 配置失败恢复策略：失败后 1 分钟重启两次，每天一次
//   - 如果服务已存在，安装会失败
//
// 命令行使用
//
//	main.exe install
//
// 示例
//
//	svc := server.DefaultWinSVC(myHandler)
//	svc.Name = "MyService"
//	svc.DisplayName = "我的服务"
//	svc.Install()
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

// Remove 卸载 Windows 服务
//
// 从系统中删除已安装的 Windows 服务
//
// 注意事项
//   - 必须以管理员身份运行
//   - 如果不是管理员，会自动尝试提升权限
//   - 如果服务正在运行，需要先停止服务
//   - 如果服务不存在，卸载会失败
//
// 命令行使用
//
//	main.exe remove
//
// 示例
//
//	svc := server.DefaultWinSVC(myHandler)
//	svc.Name = "MyService"
//	svc.Remove()
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
