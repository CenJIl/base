//go:build !windows

package logger

func isWindowsService() bool {
	return false
}
