package band

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type AgentClient struct {
	APIKey  string
	AgentID string
	ChatID  string
	APIBase string
	WSBase  string
	HTTP    *http.Client
}

func NewAgentClient(chatID, apiKey, agentID string) *AgentClient {
	return &AgentClient{
		APIKey:  firstNonEmpty(apiKey, os.Getenv("BAND_API_KEY"), os.Getenv("BAND_TOKEN")),
		AgentID: firstNonEmpty(agentID, os.Getenv("BAND_AGENT_ID")),
		ChatID:  firstNonEmpty(chatID, os.Getenv("BAND_CHAT_ID"), os.Getenv("BAND_ROOM_ID")),
		APIBase: envOr("BAND_API_BASE", DefaultAPIBase),
		WSBase:  envOr("BAND_WS_BASE", DefaultWSBase),
		HTTP:    &http.Client{Timeout: 45 * time.Second},
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

type agentProfile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type meResponse struct {
	Agent agentProfile `json:"agent"`
}

func (c *AgentClient) GetMe(ctx context.Context) (*agentProfile, error) {
	var resp meResponse
	if err := c.doJSON(ctx, http.MethodGet, "/me", nil, &resp); err != nil {
		return nil, err
	}
	if resp.Agent.ID != "" {
		return &resp.Agent, nil
	}
	return nil, fmt.Errorf("band: empty agent profile")
}

type bandMessage struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

type messageEnvelope struct {
	Message bandMessage `json:"message"`
}

func (c *AgentClient) NextMessage(ctx context.Context, chatID string) (*bandMessage, error) {
	if chatID == "" {
		chatID = c.ChatID
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.APIBase+"/chats/"+chatID+"/messages/next", nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, ErrNoMessages
	}
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("band next message %s: %s", resp.Status, string(body))
	}

	var env messageEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return nil, err
	}
	if env.Message.ID == "" {
		return nil, fmt.Errorf("band: message missing id")
	}
	return &env.Message, nil
}

func (c *AgentClient) MarkProcessing(ctx context.Context, chatID, messageID string) error {
	return c.postEmpty(ctx, chatID, messageID, "processing")
}

func (c *AgentClient) MarkProcessed(ctx context.Context, chatID, messageID string) error {
	return c.postEmpty(ctx, chatID, messageID, "processed")
}

func (c *AgentClient) MarkFailed(ctx context.Context, chatID, messageID, reason string) error {
	path := fmt.Sprintf("/chats/%s/messages/%s/failed", chatID, messageID)
	body := map[string]interface{}{"error": reason}
	return c.doJSON(ctx, http.MethodPost, path, body, nil)
}

func (c *AgentClient) PostEvent(ctx context.Context, chatID, content, messageType string, metadata map[string]interface{}) error {
	if chatID == "" {
		chatID = c.ChatID
	}
	payload := map[string]interface{}{
		"event": map[string]interface{}{
			"content":      content,
			"message_type": messageType,
		},
	}
	if metadata != nil {
		payload["event"].(map[string]interface{})["metadata"] = metadata
	}
	return c.doJSON(ctx, http.MethodPost, "/chats/"+chatID+"/events", payload, nil)
}

func (c *AgentClient) PostLog(ctx context.Context, chatID, content string) error {
	return c.PostEvent(ctx, chatID, content, "thought", nil)
}

func (c *AgentClient) PostRelease(ctx context.Context, chatID string, manifest map[string]interface{}) error {
	return c.PostEvent(ctx, chatID, "release published", "tool_result", manifest)
}

func (c *AgentClient) postEmpty(ctx context.Context, chatID, messageID, action string) error {
	path := fmt.Sprintf("/chats/%s/messages/%s/%s", chatID, messageID, action)
	return c.doJSON(ctx, http.MethodPost, path, map[string]interface{}{}, nil)
}

func (c *AgentClient) doJSON(ctx context.Context, method, path string, body interface{}, out interface{}) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}
	url := c.APIBase + path
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return err
	}
	c.setHeaders(req)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("band api %s %s: %s", method, resp.Status, string(b))
	}
	if out != nil && resp.StatusCode != http.StatusNoContent {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func (c *AgentClient) setHeaders(req *http.Request) {
	if c.APIKey != "" {
		req.Header.Set("X-API-Key", c.APIKey)
	}
}

// Client is an alias wrapper kept for orchestrator compatibility.
type Client = AgentClient

func NewClient(chatID, apiKey string) *AgentClient {
	return NewAgentClient(chatID, apiKey, "")
}

func (c *AgentClient) Listen(ctx context.Context, handler func(Event) error) error {
	me, err := c.GetMe(ctx)
	if err != nil {
		return err
	}
	if c.AgentID == "" {
		c.AgentID = me.ID
	}

	for {
		msg, err := c.NextMessage(ctx, c.ChatID)
		if err == ErrNoMessages {
			break
		}
		if err != nil {
			return err
		}
		task := Task{ChatID: c.ChatID, MessageID: msg.ID, Content: msg.Content}
		if err := c.processTask(ctx, task, handler); err != nil {
			return err
		}
	}

	ws := newPhoenixConn(c)
	return ws.run(ctx, handler)
}

func (c *AgentClient) processTask(ctx context.Context, task Task, handler func(Event) error) error {
	if err := c.MarkProcessing(ctx, task.ChatID, task.MessageID); err != nil {
		return err
	}
	err := handler(task.ToEvent())
	if err != nil {
		_ = c.MarkFailed(ctx, task.ChatID, task.MessageID, err.Error())
		return err
	}
	return c.MarkProcessed(ctx, task.ChatID, task.MessageID)
}
