package docker

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const composeLockPath = "/tmp/golang-app-test.compose.lock"

func lock(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			// write pid for observability
			_, _ = fmt.Fprintf(f, "%d\n", os.Getpid())
			_ = f.Close()
			return nil
		}
		if !os.IsExist(err) {
			return fmt.Errorf("lock: unexpected error: %w", err)
		}
		if time.Now().After(deadline) {
			// read current owner for diagnostics
			b, _ := os.ReadFile(path)
			return fmt.Errorf("lock: timeout acquiring %s (owner pid: %s)", path, strings.TrimSpace(string(b)))
		}
		time.Sleep(300 * time.Millisecond)
	}
}

func unlock(path string) {
	_ = os.Remove(path)
}
