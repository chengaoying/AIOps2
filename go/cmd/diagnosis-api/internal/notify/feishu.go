package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type FeishuChannel struct {
	webhookURL string
	httpClient *http.Client
}

type FeishuMsg struct {
	MsgType string `json:"msg_type"`
	Content struct {
		Text string `json:"text"`
	} `json:"content"`
}

func NewFeishuChannel(webhookURL string) *FeishuChannel {
	return &FeishuChannel{
		webhookURL: webhookURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *FeishuChannel) Send(ctx context.Context, title, content string) error {
	msg := FeishuMsg{MsgType: "text"}
	msg.Content.Text = fmt.Sprintf("%s\n\n%s", title, content)

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
		return fmt.Errorf("feishu api returned %d", resp.StatusCode)
	}
	return nil
}
