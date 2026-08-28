//go:build windows

package update

import "os/exec"

func Restart(executable string, args, env []string) error {
	cmd := exec.Command(executable)
	cmd.Args = args
	cmd.Env = env
	return cmd.Start()
}
