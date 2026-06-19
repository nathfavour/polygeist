package ipc

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
)

type Request struct {
	Op string `json:"op"`
}

type HealthReport struct {
	Status    string            `json:"status"`
	Phase     string            `json:"phase"`
	Sockets   map[string]string `json:"sockets"`
	Reachable map[string]bool   `json:"reachable"`
}

type Server struct {
	mu    sync.RWMutex
	phase string
}

func NewServer() *Server {
	return &Server{phase: "idle"}
}

func (s *Server) SetPhase(phase string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.phase = phase
}

func (s *Server) Start(ctx context.Context) error {
	path := SocketPath()
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	_ = os.Remove(path)

	l, err := net.Listen("unix", path)
	if err != nil {
		return fmt.Errorf("polygeist ipc listen: %w", err)
	}
	_ = os.Chmod(path, 0o600)

	go func() {
		<-ctx.Done()
		l.Close()
	}()

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					continue
				}
			}
			go s.handle(conn)
		}
	}()
	return nil
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	var req Request
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		return
	}

	switch req.Op {
	case "HEALTH", "STATUS":
		_ = json.NewEncoder(conn).Encode(s.health())
	default:
		_ = json.NewEncoder(conn).Encode(map[string]string{"status": "ERROR", "message": "unknown op"})
	}
}

func (s *Server) health() HealthReport {
	s.mu.RLock()
	phase := s.phase
	s.mu.RUnlock()

	sockets := map[string]string{
		"polygeist": SocketPath(),
		"anyisland": envOr("ANYISLAND_SOCKET", envOr("ANYISLAND_IPC_SOCK", "")),
		"vibeauracle": envOr("VIBEAURA_SOCKET", envOr("VIBEIPC_SOCK", "")),
	}
	reachable := map[string]bool{}
	for name, path := range sockets {
		if path == "" {
			continue
		}
		reachable[name] = dialOK(path)
	}

	return HealthReport{
		Status:    "ok",
		Phase:     phase,
		Sockets:   sockets,
		Reachable: reachable,
	}
}

func dialOK(path string) bool {
	conn, err := net.Dial("unix", path)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
