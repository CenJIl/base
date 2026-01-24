//go:build !windows

package server

import (
	"context"
	"time"

	"github.com/CenJIl/base/common"
)

type WinSVC struct {
	Name         string
	DisplayName  string
	Description  string
	Log          common.Logger
	ShutdownWait time.Duration
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
