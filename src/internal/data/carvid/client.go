package carvid

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"data-agent/pkg/logger"

	"go.uber.org/zap"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	log        *zap.SugaredLogger
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		log: logger.GetLogger(),
	}
}

func (c *Client) ResolveIdentity(ctx context.Context, platform, id string) (IdentityInfo, error) {
	url := fmt.Sprintf("%s/identity/resolve?platform=%s&id=%s", c.baseURL, platform, id)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return IdentityInfo{}, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Errorw("Failed to resolve identity",
			"error", err,
			"platform", platform,
			"id", id)
		return IdentityInfo{}, err
	}
	defer resp.Body.Close()

	var result IdentityInfo
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return IdentityInfo{}, err
	}

	return result, nil
}
