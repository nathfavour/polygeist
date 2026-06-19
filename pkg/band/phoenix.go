package band

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type phoenixConn struct {
	client *AgentClient
	ref    int
}

func newPhoenixConn(c *AgentClient) *phoenixConn {
	return &phoenixConn{client: c, ref: 1}
}

func (p *phoenixConn) nextRef() string {
	p.ref++
	return fmt.Sprintf("%d", p.ref)
}

func (p *phoenixConn) dialURL() (string, error) {
	u, err := url.Parse(p.client.WSBase)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("api_key", p.client.APIKey)
	q.Set("vsn", "2.0.0")
	if p.client.AgentID != "" {
		q.Set("agent_id", p.client.AgentID)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (p *phoenixConn) run(ctx context.Context, handler func(Event) error) error {
	wsURL, err := p.dialURL()
	if err != nil {
		return err
	}

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, http.Header{})
	if err != nil {
		return fmt.Errorf("band websocket dial: %w", err)
	}
	defer conn.Close()

	joinRef := "1"
	channels := []string{
		fmt.Sprintf("chat_room:%s", p.client.ChatID),
		fmt.Sprintf("agent_rooms:%s", p.client.AgentID),
	}
	for _, topic := range channels {
		if strings.HasSuffix(topic, ":") {
			continue
		}
		if err := p.send(conn, joinRef, topic, "phx_join", map[string]interface{}{}); err != nil {
			return err
		}
	}

	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	errCh := make(chan error, 1)
	go func() {
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				errCh <- err
				return
			}
			if err := p.handleFrame(ctx, data, handler); err != nil {
				errCh <- err
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			if err != nil {
				return fmt.Errorf("band websocket: %w", err)
			}
			return nil
		case <-heartbeat.C:
			if err := p.send(conn, "", "phoenix", "heartbeat", map[string]interface{}{}); err != nil {
				return err
			}
		}
	}
}

func (p *phoenixConn) send(conn *websocket.Conn, joinRef, topic, event string, payload map[string]interface{}) error {
	msg := []interface{}{joinRef, p.nextRef(), topic, event, payload}
	return conn.WriteJSON(msg)
}

func (p *phoenixConn) handleFrame(ctx context.Context, data []byte, handler func(Event) error) error {
	var frame []json.RawMessage
	if err := json.Unmarshal(data, &frame); err != nil {
		return nil
	}
	if len(frame) < 5 {
		return nil
	}

	var topic, event string
	_ = json.Unmarshal(frame[2], &topic)
	_ = json.Unmarshal(frame[3], &event)

	switch event {
	case "phx_reply", "heartbeat", "phx_close":
		return nil
	case "message_created":
		var payload struct {
			Message struct {
				ID      string `json:"id"`
				Content string `json:"content"`
				ChatID  string `json:"chat_id"`
			} `json:"message"`
		}
		if err := json.Unmarshal(frame[4], &payload); err != nil {
			// Some payloads nest differently; try flat decode.
			var flat struct {
				ID      string `json:"id"`
				Content string `json:"content"`
			}
			if err2 := json.Unmarshal(frame[4], &flat); err2 != nil {
				return nil
			}
			payload.Message.ID = flat.ID
			payload.Message.Content = flat.Content
		}
		chatID := payload.Message.ChatID
		if chatID == "" {
			chatID = strings.TrimPrefix(topic, "chat_room:")
		}
		if payload.Message.ID == "" || payload.Message.Content == "" {
			return nil
		}
		task := Task{
			ChatID:    chatID,
			MessageID: payload.Message.ID,
			Content:   payload.Message.Content,
		}
		return p.client.processTask(ctx, task, handler)
	default:
		return nil
	}
}
