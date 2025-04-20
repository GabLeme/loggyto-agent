package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log-agent/internal/logentry"
	"net/http"
)

type Sender struct {
	endpoint  string
	apiKey    string
	apiSecret string
	client    *http.Client
}

func NewSender(cfg Config) *Sender {
	return &Sender{
		endpoint:  cfg.Endpoint,
		apiKey:    cfg.APIKey,
		apiSecret: cfg.APISecret,
		client: &http.Client{
			Timeout: 15 * 1_000_000_000, // 15s
		},
	}
}

func (s *Sender) Send(entry logentry.LogEntry) error {
	payload, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	req, err := http.NewRequest("POST", s.endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", s.apiKey)
	req.Header.Set("x-api-secret", s.apiSecret)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
