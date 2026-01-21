//go:build !windows

package server

import (
	"context"
)

type Logger interface {
	Infof(format string, v ...any)
	Errorf(format string, v ...any)
}

type defaultLog struct{}

func (l *defaultLog) Infof(f string, v ...any)  {}
func (l *defaultLog) Errorf(f string, v ...any) {}

type WinSVC struct {
	Name         string
	DisplayName  string
	Description  string
	Log          Logger
	ShutdownWait int
	Handler      func(ctx context.Context) error
}

func (w *WinSVC) String() string {
	panic("WinSVC.String 不支持非 Windows 平台")
}

func DefaultWinSVC(handler func(ctx context.Context) error) *WinSVC {
	panic("DefaultWinSVC 不支持非 Windows 平台")
}

func (w *WinSVC) Run() {
	panic("WinSVC.Run 不支持非 Windows 平台")
}

func (w *WinSVC) Execute(args []string, r <-chan interface{}, changes chan<- interface{}) (bool, uint32) {
	panic("WinSVC.Execute 不支持非 Windows 平台")
}

func (w *WinSVC) Install() {
	panic("WinSVC.Install 不支持非 Windows 平台")
}

func (w *WinSVC) Remove() {
	panic("WinSVC.Remove 不支持非 Windows 平台")
}
