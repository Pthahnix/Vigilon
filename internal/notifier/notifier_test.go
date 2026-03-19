package notifier

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuditLog(t *testing.T) {
	dir := t.TempDir()
	n := &Notifier{LogPath: dir, Wall: false}
	n.Log("test-event", "hyf", "test message")

	data, err := os.ReadFile(filepath.Join(dir, "audit.log"))
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if !strings.Contains(string(data), "test message") {
		t.Errorf("log missing message: %s", data)
	}
	if !strings.Contains(string(data), "hyf") {
		t.Errorf("log missing user: %s", data)
	}
}
