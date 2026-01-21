//go:build windows

package server

import (
	"context"
	"fmt"
	"testing"
)

func TestDefault(t *testing.T) {
	fmt.Printf("%v", DefaultWinSVC(GetName))
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
