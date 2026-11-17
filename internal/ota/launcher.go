package ota

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const healthTimeout = 15 * time.Second
const (
	HealthFileEnv  = "WAYCORD_HEALTH_FILE"
	HealthTokenEnv = "WAYCORD_HEALTH_TOKEN"
)

// MarkHealthy should only be called after the first GUI frame has been drawn.
// It does not authenticate a process; it is a startup crash-detection handshake.
func MarkHealthy() error {
	path, token := os.Getenv(HealthFileEnv), os.Getenv(HealthTokenEnv)
	if path == "" && token == "" {
		return nil
	}
	if path == "" || len(token) != 64 {
		return errors.New("incomplete launcher health challenge")
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	if _, err = f.WriteString(token); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func (s Store) launchVersion(ctx context.Context, v string, env []string) error {
	path, err := s.Binary(v)
	if err != nil {
		return err
	}
	if _, err = os.Stat(path); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, path)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Launch executes the pending build and promotes it only after it signals that
// its first frame rendered. On a crash/timeout, the old version remains active.
func (s Store) Launch(ctx context.Context) error {
	st, _, _, err := s.Load()
	if err != nil {
		return err
	}
	if st.Pending == "" {
		if st.Current == "" {
			return errors.New("nothing installed; run 'install' first")
		}
		return s.launchVersion(ctx, st.Current, nil)
	}
	if err = s.launchPending(ctx, st.Pending); err != nil {
		// A failure after promotion is an ordinary application exit, not an OTA failure.
		latest, _, _, loadErr := s.Load()
		if loadErr != nil {
			return loadErr
		}
		if latest.Pending == "" {
			return err
		}
		if rollbackErr := s.Rollback(); rollbackErr != nil {
			return fmt.Errorf("update failed: %v; rollback failed: %w", err, rollbackErr)
		}
		if st.Current == "" {
			return fmt.Errorf("first launch failed: %w", err)
		}
		fmt.Fprintln(os.Stderr, "Waycord: update failed; launching previous version:", err)
		return s.launchVersion(ctx, st.Current, nil)
	}
	return nil
}

func (s Store) launchPending(ctx context.Context, v string) error {
	path, err := s.Binary(v)
	if err != nil {
		return err
	}
	if _, err = os.Stat(path); err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Join(s.Root, "run"), 0700); err != nil {
		return err
	}
	token := make([]byte, 32)
	if _, err = rand.Read(token); err != nil {
		return err
	}
	secret := hex.EncodeToString(token)
	ack := filepath.Join(s.Root, "run", "health-"+secret)
	defer os.Remove(ack)
	cmd := exec.CommandContext(ctx, path)
	cmd.Env = append(os.Environ(), HealthFileEnv+"="+ack, HealthTokenEnv+"="+secret)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	ticker := time.NewTicker(125 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.NewTimer(healthTimeout)
	defer deadline.Stop()
	for {
		select {
		case err = <-done:
			return fmt.Errorf("application exited before health check: %w", errOrExited(err))
		case <-ticker.C:
			data, readErr := os.ReadFile(ack)
			if readErr == nil && strings.TrimSpace(string(data)) == secret {
				if err = s.Promote(); err != nil {
					cmd.Process.Kill()
					<-done
					return err
				}
				return <-done // keep the launcher running until the app exits
			}
		case <-deadline.C:
			cmd.Process.Kill()
			<-done
			return errors.New("application failed startup health check in time")
		case <-ctx.Done():
			cmd.Process.Kill()
			<-done
			return ctx.Err()
		}
	}
}

func errOrExited(err error) error {
	if err != nil {
		return err
	}
	return errors.New("exited cleanly before first frame")
}
