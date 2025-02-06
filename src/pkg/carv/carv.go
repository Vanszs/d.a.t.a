package carv

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type CarvConfig struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
}

type Client struct {
	APIKey  string
	BaseURL string
}

type Balance struct {
	Amount       float64
	Network      string
	ContractAddr string
}

func NewClient(apiKey string, baseURL string) *Client {
	return &Client{
		APIKey:  apiKey,
		BaseURL: baseURL,
	}
}

func (d *Client) GetBalanceByDiscordID(
	ctx context.Context,
	userID string,
	chainName string,
	tokenTicker string,
) (*Balance, error) {
	url := fmt.Sprintf(
		"%s/user_balance_by_discord_id?discord_user_id=%s&chain_name=%s&token_ticker=%s",
		d.BaseURL,
		userID,
		chainName,
		tokenTicker,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Authorization", d.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var balanceResponse struct {
		Data struct {
			Balance string `json:"balance"`
		} `json:"data"`
		Code    int    `json:"code"`
		Message string `json:"msg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&balanceResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	floatValue, err := strconv.ParseFloat(balanceResponse.Data.Balance, 64)
	if err != nil {
		return nil, err
	}
	return &Balance{
		Amount:  floatValue,
		Network: chainName,
	}, nil
}
