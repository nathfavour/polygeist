package band

import (
	"errors"
	"time"
)

const (
	DefaultAPIBase = "https://app.band.ai/api/v1/agent"
	DefaultWSBase  = "wss://app.band.ai/api/v1/socket/websocket"
)

var ErrNoMessages = errors.New("no messages")

// Task is a Band chat message polygeist should process.
type Task struct {
	ChatID    string
	MessageID string
	Content   string
	Metadata  map[string]interface{}
}

// Event is kept for local/offline compatibility.
type Event struct {
	EventID   string                 `json:"event_id"`
	RoomID    string                 `json:"room_id"`
	SenderID  string                 `json:"sender_id"`
	Payload   string                 `json:"payload"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp int64                  `json:"timestamp"`
}

func (e Event) WorkDir() string {
	if e.Metadata == nil {
		return "."
	}
	if v, ok := e.Metadata["work_dir"].(string); ok && v != "" {
		return v
	}
	return "."
}

func (e Event) VerifyCommand() string {
	if e.Metadata == nil {
		return ""
	}
	if v, ok := e.Metadata["verify_command"].(string); ok {
		return v
	}
	return ""
}

func (e Event) PackageName() string {
	if e.Metadata == nil {
		return "polygeist"
	}
	if v, ok := e.Metadata["package"].(string); ok && v != "" {
		return v
	}
	return "polygeist"
}

func (e Event) Version() string {
	if e.Metadata == nil {
		return "dev"
	}
	if v, ok := e.Metadata["version"].(string); ok && v != "" {
		return v
	}
	return "dev"
}

func TaskFromEvent(evt Event) Task {
	return Task{
		ChatID:    evt.RoomID,
		MessageID: evt.EventID,
		Content:   evt.Payload,
		Metadata:  evt.Metadata,
	}
}

func (t Task) WorkDir() string {
	if t.Metadata == nil {
		return "."
	}
	if v, ok := t.Metadata["work_dir"].(string); ok && v != "" {
		return v
	}
	return "."
}

func (t Task) VerifyCommand() string {
	if t.Metadata == nil {
		return ""
	}
	if v, ok := t.Metadata["verify_command"].(string); ok {
		return v
	}
	return ""
}

func (t Task) PackageName() string {
	if t.Metadata == nil {
		return "polygeist"
	}
	if v, ok := t.Metadata["package"].(string); ok && v != "" {
		return v
	}
	return "polygeist"
}

func (t Task) Version() string {
	if t.Metadata == nil {
		return "dev"
	}
	if v, ok := t.Metadata["version"].(string); ok && v != "" {
		return v
	}
	return "dev"
}

func (t Task) ToEvent() Event {
	return Event{
		EventID:   t.MessageID,
		RoomID:    t.ChatID,
		Payload:   t.Content,
		Metadata:  t.Metadata,
		Timestamp: time.Now().Unix(),
	}
}
