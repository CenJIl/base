//go:build windows

package logger

import "golang.org/x/sys/windows/svc"

func isWindowsService() bool {
	isService, err := svc.IsWindowsService()
	return err == nil && isService
}
