package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/temo927/go-msg-dispatcher/internal/domain"
	"github.com/temo927/go-msg-dispatcher/internal/infra/log"
)

type Config struct {
	URL          string
	AuthHeader   string
	AuthValue    string
	AcceptAny2xx bool
	Timeout      time.Duration
}

type Client struct {
	http *http.Client
	cfg  Config
}

func NewClient(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second 
	}
	return &Client{
		http: &http.Client{Timeout: cfg.Timeout},
		cfg:  cfg,
	}
}

type webhookPayload struct {
	To      string `json:"to"`
	Content string `json:"content"`
}

type webhookResponse struct {
	Message   string `json:"message"`
	MessageID string `json:"messageId"`
}

func (c *Client) Send(ctx context.Context, msg domain.Message) (string, error) {
	body, err := json.Marshal(webhookPayload{
		To:      msg.ToPhone,
		Content: msg.Content,
	})
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.URL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.cfg.AuthHeader != "" && c.cfg.AuthValue != "" {
		req.Header.Set(c.cfg.AuthHeader, c.cfg.AuthValue)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("http send: %w", err)
	}
	defer resp.Body.Close()

	if !c.acceptedStatus(resp.StatusCode) {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		log.Logger.Error("webhook non-2xx", "status", resp.StatusCode, "body", string(b))
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(b))
	}

	var res webhookResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if res.MessageID == "" {
		return "", fmt.Errorf("missing messageId in webhook response")
	}

	log.Logger.Info("webhook accepted", "msg_id", msg.ID, "provider_message_id", res.MessageID)
	return res.MessageID, nil
}

func (c *Client) acceptedStatus(code int) bool {
	if code == http.StatusAccepted { // 202
		return true
	}
	if c.cfg.AcceptAny2xx {
		return code >= 200 && code < 300
	}
	return false
}
