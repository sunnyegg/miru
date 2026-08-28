//go:build !windows

package update

import "syscall"

func Restart(executable string, args, env []string) error {
	return syscall.Exec(executable, args, env)
}
