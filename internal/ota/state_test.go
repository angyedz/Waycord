package ota

import (
	"os"
	"testing"
)

func TestStateStagePromoteRollback(t *testing.T) {
	s := Store{Root: t.TempDir()}
	if err := s.Stage("v0.1.0"); err != nil {
		t.Fatal(err)
	}
	st, _, _, err := s.Load()
	if err != nil || st.Pending != "v0.1.0" {
		t.Fatalf("stage: %+v %v", st, err)
	}
	if err := s.Promote(); err != nil {
		t.Fatal(err)
	}
	st, _, _, _ = s.Load()
	if st.Current != "v0.1.0" || st.Pending != "" {
		t.Fatalf("promote %+v", st)
	}
	if err := s.Stage("v0.2.0"); err != nil {
		t.Fatal(err)
	}
	if err := s.Rollback(); err != nil {
		t.Fatal(err)
	}
	st, _, _, _ = s.Load()
	if st.Current != "v0.1.0" || st.Pending != "" {
		t.Fatalf("rollback %+v", st)
	}
	if err := s.Stage("../../evil"); err == nil {
		t.Fatal("unsafe version accepted")
	}
}
func TestDoubleSlotRecoversPreviousState(t *testing.T) {
	s := Store{Root: t.TempDir()}
	if err := s.Save(State{Current: "v1.0.0"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(State{Current: "v1.1.0"}); err != nil {
		t.Fatal(err)
	}
	_, _, slot, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.statePath(slot), []byte("partial"), 0600); err != nil {
		t.Fatal(err)
	}
	st, _, _, err := s.Load()
	if err != nil || st.Current != "v1.0.0" {
		t.Fatalf("failed fallback %+v %v", st, err)
	}
}
func TestBothCorruptStateFailsClosed(t *testing.T) {
	s := Store{Root: t.TempDir()}
	if err := os.MkdirAll(s.Root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.statePath(0), []byte("garbage"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := s.Load(); err == nil {
		t.Fatal("corrupt state should not be treated as fresh installation")
	}
}
