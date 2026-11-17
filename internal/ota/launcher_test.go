package ota

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestMain makes the test binary usable as a real, cross-platform child process.
func TestMain(m *testing.M) {
	if os.Getenv("WAYCORD_TEST_CHILD") == "1" {
		if err := MarkHealthy(); err != nil {
			os.Exit(2)
		}
		if os.Getenv(HealthFileEnv) != "" && os.Getenv("WAYCORD_TEST_FREEZE") == "1" {
			time.Sleep(time.Second)
			os.Exit(0)
		}
		started := time.Now()
		for time.Since(started) < 500*time.Millisecond {
			if os.Getenv("WAYCORD_TEST_CRASH") == "1" && os.Getenv(HealthFileEnv) != "" && time.Since(started) > 150*time.Millisecond {
				os.Exit(3)
			}
			if os.Getenv(HealthFileEnv) != "" {
				if err := PulseHealthy(); err != nil {
					os.Exit(4)
				}
			}
			time.Sleep(20 * time.Millisecond)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func pendingTestStore(t *testing.T) Store {
	t.Helper()
	s := Store{Root: t.TempDir()}
	if err := s.Save(State{Current: "v1.0.0", Pending: "v1.1.0"}); err != nil {
		t.Fatal(err)
	}
	target, err := s.Binary("v1.1.0")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
		t.Fatal(err)
	}
	source, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, data, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WAYCORD_TEST_CHILD", "1")
	return s
}

func TestPendingCrashAfterReadyDoesNotPromote(t *testing.T) {
	s := pendingTestStore(t)
	t.Setenv("WAYCORD_TEST_CRASH", "1")
	if err := s.launchPendingWithPolicy(context.Background(), "v1.1.0", 2*time.Second, 300*time.Millisecond, 10*time.Millisecond, 200*time.Millisecond); err == nil {
		t.Fatal("crash during observation accepted")
	}
	st, _, _, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if st.Current != "v1.0.0" || st.Pending != "v1.1.0" {
		t.Fatalf("premature promotion: %+v", st)
	}
}

func TestPendingPromotesAfterReadyAndStability(t *testing.T) {
	s := pendingTestStore(t)
	if err := s.launchPendingWithPolicy(context.Background(), "v1.1.0", 2*time.Second, 100*time.Millisecond, 10*time.Millisecond, 200*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	st, _, _, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if st.Current != "v1.1.0" || st.Previous != "v1.0.0" || st.Pending != "" {
		t.Fatalf("bad promotion: %+v", st)
	}
}

func TestHealthRecordRejectsTampering(t *testing.T) {
	token := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	if !validHealthRecord([]byte(token+"..."), token) {
		t.Fatal("valid heartbeat record rejected")
	}
	for _, value := range []string{token[:63], token + "x", "x" + token} {
		if validHealthRecord([]byte(value), token) {
			t.Fatalf("invalid heartbeat record accepted: %q", value)
		}
	}
}

func TestReplaceEnvironmentOverridesDuplicateHealthChallenge(t *testing.T) {
	got := replaceEnvironment(
		[]string{"PATH=/bin", HealthFileEnv + "=/stale", HealthTokenEnv + "=stale", "OTHER=value"},
		HealthFileEnv+"=/fresh", HealthTokenEnv+"=fresh",
	)
	want := []string{"PATH=/bin", "OTHER=value", HealthFileEnv + "=/fresh", HealthTokenEnv + "=fresh"}
	if len(got) != len(want) {
		t.Fatalf("unexpected env length: %q", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected env at %d: got %q want %q", i, got[i], want[i])
		}
	}
}

func TestPendingFrozenAfterReadyDoesNotPromote(t *testing.T) {
	s := pendingTestStore(t)
	t.Setenv("WAYCORD_TEST_FREEZE", "1")
	if err := s.launchPendingWithPolicy(context.Background(), "v1.1.0", 2*time.Second, 300*time.Millisecond, 10*time.Millisecond, 120*time.Millisecond); err == nil {
		t.Fatal("frozen application accepted")
	}
	st, _, _, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if st.Current != "v1.0.0" || st.Pending != "v1.1.0" {
		t.Fatalf("frozen candidate changed install state: %+v", st)
	}
}

func TestLaunchRollsBackAfterReadyCrash(t *testing.T) {
	s := pendingTestStore(t)
	t.Setenv("WAYCORD_TEST_CRASH", "1")
	previous, err := s.Binary("v1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(previous), 0700); err != nil {
		t.Fatal(err)
	}
	source, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(previous, data, 0700); err != nil {
		t.Fatal(err)
	}
	if err := s.Launch(context.Background()); err != nil {
		t.Fatal(err)
	}
	st, _, _, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if st.Current != "v1.0.0" || st.Pending != "" {
		t.Fatalf("rollback failed: %+v", st)
	}
}
