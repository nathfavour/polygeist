package band

import "encoding/json"

// Event represents structured JSON data from the Band.ai network.
type Event struct {
	EventID   string                 `json:"event_id"`
	RoomID    string                 `json:"room_id"`
	SenderID  string                 `json:"sender_id"`
	Payload   string                 `json:"payload"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp int64                  `json:"timestamp"`
}

// OutboundMessage is sent to a Band room via REST egress.
type OutboundMessage struct {
	RoomID  string `json:"room_id"`
	Content string `json:"content"`
	Type    string `json:"type,omitempty"`
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

func (m OutboundMessage) JSON() ([]byte, error) {
	return json.Marshal(m)
}
