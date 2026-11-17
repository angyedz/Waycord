package attachments

import (
	"bytes"
	"testing"
)

func TestReadLimitsSizeAndKeepsOnlyFileName(t *testing.T) {
	file, err := Read("folder/example.txt", bytes.NewReader([]byte("hello")))
	if err != nil || file.Name != "example.txt" || string(file.Data) != "hello" {
		t.Fatalf("unexpected file: %+v, %v", file, err)
	}
	if _, err := Read("large", bytes.NewReader(make([]byte, MaxSize+1))); err == nil {
		t.Fatal("accepted oversized attachment")
	}
}
