package band

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

const (
	DefaultWSURL  = "wss://api.band.ai/v1/ws"
	DefaultAPIURL = "https://api.band.ai/v1"
)

type Client struct {
	WSURL     string
	APIURL    string
	Token     string
	RoomID    string
	HTTP      *http.Client
	dialer    websocket.Dialer
}

func NewClient(roomID, token string) *Client {
	return &Client{
		WSURL:  envOr("BAND_WS_URL", DefaultWSURL),
		APIURL: envOr("BAND_API_URL", DefaultAPIURL),
		Token:  token,
		RoomID: roomID,
		HTTP: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func (c *Client) Listen(ctx context.Context, handler func(Event) error) error {
	header := http.Header{}
	if c.Token != "" {
		header.Set("Authorization", "Bearer "+c.Token)
	}

	conn, _, err := c.dialer.DialContext(ctx, c.WSURL, header)
	if err != nil {
		return fmt.Errorf("websocket dial: %w", err)
	}
	defer conn.Close()

	if c.RoomID != "" {
		sub := map[string]string{"op": "subscribe", "room_id": c.RoomID}
		if err := conn.WriteJSON(sub); err != nil {
			return fmt.Errorf("subscribe: %w", err)
		}
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var evt Event
		if err := conn.ReadJSON(&evt); err != nil {
			return fmt.Errorf("read event: %w", err)
		}
		if c.RoomID != "" && evt.RoomID != "" && evt.RoomID != c.RoomID {
			continue
		}
		if err := handler(evt); err != nil {
			return err
		}
	}
}

func (c *Client) PostLog(ctx context.Context, roomID, content string) error {
	return c.postMessage(ctx, OutboundMessage{
		RoomID:  roomID,
		Content: content,
		Type:    "log",
	})
}

func (c *Client) PostRelease(ctx context.Context, roomID string, manifest map[string]interface{}) error {
	data, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	return c.postMessage(ctx, OutboundMessage{
		RoomID:  roomID,
		Content: string(data),
		Type:    "release",
	})
}

func (c *Client) postMessage(ctx context.Context, msg OutboundMessage) error {
	if msg.RoomID == "" {
		msg.RoomID = c.RoomID
	}
	body, err := msg.JSON()
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.APIURL+"/rooms/"+msg.RoomID+"/messages", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("band api %s: %s", resp.Status, string(b))
	}
	return nil
}
