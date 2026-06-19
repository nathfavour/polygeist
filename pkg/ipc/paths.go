package ipc

import (
	"os"
	"path/filepath"
)

// SocketPath returns the polygeist status UDS path.
func SocketPath() string {
	if v := os.Getenv("POLYGEIST_SOCKET"); v != "" {
		return v
	}
	if run := os.Getenv("AGENTIC_RUN_DIR"); run != "" {
		return filepath.Join(run, "polygeist.sock")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".polygeist", "polygeist.sock")
}

// RunDir returns the shared runtime directory for agent UDS sockets.
func RunDir() string {
	if run := os.Getenv("AGENTIC_RUN_DIR"); run != "" {
		return run
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".polygeist", "run")
}
