package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// UDMClient handles communication with UDM
type UDMClient struct {
	baseURL string
	client  *http.Client
	logger  *zap.Logger
}

// NewUDMClient creates a new UDM client
func NewUDMClient(baseURL string, timeout time.Duration, logger *zap.Logger) *UDMClient {
	return &UDMClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: timeout,
		},
		logger: logger,
	}
}

// AuthenticationInfo represents authentication information request to UDM
type AuthenticationInfo struct {
	SUPI                  string `json:"supi"`
	ServingNetworkName    string `json:"servingNetworkName"`
	ResynchronizationInfo *struct {
		RAND string `json:"rand"`
		AUTS string `json:"auts"`
	} `json:"resynchronizationInfo,omitempty"`
}

// AuthenticationVector represents a 5G AKA authentication vector
type AuthenticationVector struct {
	RAND  string `json:"rand"`  // Random challenge (hex)
	AUTN  string `json:"autn"`  // Authentication token (hex)
	HXRES string `json:"hxres"` // Expected response (hex)
	KAUSF string `json:"kausf"` // Key for AUSF (hex)
}

// AuthenticationInfoResult represents the authentication response from UDM
type AuthenticationInfoResult struct {
	AuthType             string                `json:"authType"` // "5G_AKA" or "EAP_AKA_PRIME"
	AuthenticationVector *AuthenticationVector `json:"authenticationVector,omitempty"`
}

// GenerateAuthData requests UDM to generate authentication data
func (c *UDMClient) GenerateAuthData(ctx context.Context, authInfo *AuthenticationInfo) (*AuthenticationInfoResult, error) {
	url := fmt.Sprintf("%s/nudm-ueau/v1/supi/%s/security-information/generate-auth-data",
		c.baseURL, authInfo.SUPI)

	body, err := json.Marshal(authInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	c.logger.Debug("Requesting auth data from UDM",
		zap.String("supi", authInfo.SUPI),
		zap.String("url", url),
	)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("UDM returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result AuthenticationInfoResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Debug("Received auth data from UDM",
		zap.String("supi", authInfo.SUPI),
		zap.String("auth_type", result.AuthType),
	)

	return &result, nil
}

// ConfirmAuth confirms authentication result with UDM
func (c *UDMClient) ConfirmAuth(ctx context.Context, supi string, authEvent map[string]interface{}) error {
	url := fmt.Sprintf("%s/nudm-ueau/v1/supi/%s/auth-events", c.baseURL, supi)

	body, err := json.Marshal(authEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("UDM returned status %d: %s", resp.StatusCode, string(respBody))
	}

	c.logger.Debug("Confirmed auth with UDM", zap.String("supi", supi))
	return nil
}
