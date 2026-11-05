package main

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"charm.land/ssh"
)

func TestNewServer(t *testing.T) {
	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, "host_ed25519")

	s, err := newServer("127.0.0.1:0", keyPath)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			t.Logf("server listen error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		t.Fatalf("failed to shutdown server: %v", err)
	}
}
