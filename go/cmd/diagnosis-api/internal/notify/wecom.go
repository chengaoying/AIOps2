package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type WeComChannel struct {
	webhookURL string
	httpClient *http.Client
}

type WeComMsg struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
}

func NewWeComChannel(webhookURL string) *WeComChannel {
	return &WeComChannel{
		webhookURL: webhookURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *WeComChannel) Send(ctx context.Context, title, content string) error {
	msg := WeComMsg{MsgType: "text"}
	msg.Text.Content = fmt.Sprintf("%s\n\n%s", title, content)

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.webhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("wecom api returned %d", resp.StatusCode)
	}
	return nil
}
