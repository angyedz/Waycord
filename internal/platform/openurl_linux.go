//go:build linux

package platform

import "os/exec"

func openURL(raw string) error {
	cmd := exec.Command("xdg-open", raw)
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}
