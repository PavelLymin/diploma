package exolve

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/PavelLymin/diploma/services/auth/internal/transport/http/v1/request"
)

type ExolveDelivery struct {
	client *http.Client
	config Config
}

func New(
	client *http.Client,
	config Config,
) *ExolveDelivery {
	if client == nil {
		client = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	return &ExolveDelivery{
		client: client,
		config: config,
	}
}

func (d *ExolveDelivery) SendOTP(
	ctx context.Context,
	phone, code string,
) error {
	payload := request.SMSContent{
		Number:      d.config.Sender,
		Destination: phone,
		Text:        code,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal SMS request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", d.config.BaseURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create SMS request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.config.APIKey)

	response, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("send SMS request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("send SMS request: %s", response.Status)
	}

	return nil
}
