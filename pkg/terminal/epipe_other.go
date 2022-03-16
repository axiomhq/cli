//go:build !windows

package terminal

import (
	"errors"
	"syscall"
)

func isEpipeError(err error) bool {
	return errors.Is(err, syscall.EPIPE)
}
