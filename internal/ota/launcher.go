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
const healthSilenceTimeout = 3 * time.Second
const HealthObservationWindow = 10 * time.Second
const HealthPulseInterval = time.Second
const (
	HealthFileEnv  = "WAYCORD_HEALTH_FILE"
	HealthTokenEnv = "WAYCORD_HEALTH_TOKEN"
)

// MarkHealthy should only be called after the first GUI frame has been drawn.
// It does not authenticate a process; it is a startup crash-detection handshake.
func MarkHealthy() error {
	path, token, enabled, err := healthChallenge()
	if err != nil {
		return err
	}
	if !enabled {
		return nil
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

// PulseHealthy records continued GUI-thread progress during startup observation.
func PulseHealthy() error {
	path, token, enabled, err := healthChallenge()
	if err != nil {
		return err
	}
	if !enabled {
		return nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	maxSize := int64(len(token)) + int64(HealthObservationWindow/HealthPulseInterval) + 32
	if !info.Mode().IsRegular() || info.Size() < int64(len(token)) || info.Size() >= maxSize {
		return errors.New("invalid launcher health file")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !validHealthRecord(data, token) {
		return errors.New("invalid launcher health token")
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte{'.'})
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	return err
}

func healthChallenge() (path, token string, enabled bool, err error) {
	path, token = os.Getenv(HealthFileEnv), os.Getenv(HealthTokenEnv)
	if path == "" && token == "" {
		return "", "", false, nil
	}
	decoded, decodeErr := hex.DecodeString(token)
	if path == "" || len(path) > 4096 || !filepath.IsAbs(path) || decodeErr != nil || len(decoded) != 32 {
		return "", "", false, errors.New("incomplete launcher health challenge")
	}
	return path, token, true, nil
}

func HealthChallengeEnabled() bool {
	_, _, enabled, err := healthChallenge()
	return err == nil && enabled
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

// Launch promotes a pending build only after readiness and a stability window.
// On a crash or timeout, the old version remains active.
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
	return s.launchPendingWithPolicy(ctx, v, healthTimeout, HealthObservationWindow, 125*time.Millisecond, healthSilenceTimeout)
}

func (s Store) launchPendingWithPolicy(ctx context.Context, v string, timeout, stability, poll, silence time.Duration) error {
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
	cmd.Env = replaceEnvironment(os.Environ(), HealthFileEnv+"="+ack, HealthTokenEnv+"="+secret)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	ticker := time.NewTicker(poll)
	defer ticker.Stop()
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	var stable <-chan time.Time
	var stableTimer *time.Timer
	var lastPulse time.Time
	var heartbeatSize int
	defer func() {
		if stableTimer != nil {
			stableTimer.Stop()
		}
	}()
	for {
		select {
		case err = <-done:
			return fmt.Errorf("application exited before health check: %w", errOrExited(err))
		case <-ticker.C:
			data, readErr := os.ReadFile(ack)
			if readErr == nil && validHealthRecord(data, secret) {
				if stable == nil {
					deadline.Stop()
					stableTimer = time.NewTimer(stability)
					stable = stableTimer.C
					lastPulse, heartbeatSize = time.Now(), len(data)
				} else if len(data) > heartbeatSize {
					lastPulse, heartbeatSize = time.Now(), len(data)
				}
			}
			if stable != nil && time.Since(lastPulse) > silence {
				cmd.Process.Kill()
				<-done
				return errors.New("application stopped responding during startup health check")
			}
		case <-stable:
			// Prefer a process exit already observed over promotion.
			select {
			case err = <-done:
				return fmt.Errorf("application exited during stability check: %w", errOrExited(err))
			default:
			}
			if err = s.Promote(); err != nil {
				cmd.Process.Kill()
				<-done
				return err
			}
			return <-done // keep the launcher running until the app exits
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

func validHealthRecord(data []byte, token string) bool {
	if len(data) < len(token) || string(data[:len(token)]) != token {
		return false
	}
	for _, b := range data[len(token):] {
		if b != '.' {
			return false
		}
	}
	return true
}

func replaceEnvironment(environment []string, values ...string) []string {
	replaced := make(map[string]struct{}, len(values))
	for _, value := range values {
		key, _, _ := strings.Cut(value, "=")
		replaced[key] = struct{}{}
	}
	result := make([]string, 0, len(environment)+len(values))
	for _, value := range environment {
		key, _, _ := strings.Cut(value, "=")
		if _, ok := replaced[key]; !ok {
			result = append(result, value)
		}
	}
	return append(result, values...)
}

func errOrExited(err error) error {
	if err != nil {
		return err
	}
	return errors.New("exited cleanly before first frame")
}
