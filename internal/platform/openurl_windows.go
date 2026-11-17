//go:build windows

package platform

import "os/exec"

func openURL(raw string) error {
	cmd := exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", raw)
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}
