//go:build !linux && !windows

package platform

import "errors"

func openURL(string) error { return errors.New("opening links is unsupported on this platform") }
