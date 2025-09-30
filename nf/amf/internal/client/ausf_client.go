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

// AUSFClient handles communication with AUSF
type AUSFClient struct {
	baseURL string
	client  *http.Client
	logger  *zap.Logger
}

// NewAUSFClient creates a new AUSF client
func NewAUSFClient(baseURL string, timeout time.Duration, logger *zap.Logger) *AUSFClient {
	return &AUSFClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: timeout,
		},
		logger: logger,
	}
}

// UEAuthenticationRequest represents authentication request to AUSF
type UEAuthenticationRequest struct {
	SUPI               string `json:"supiOrSuci"`
	ServingNetworkName string `json:"servingNetworkName"`
}

// UEAuthenticationResponse represents authentication response from AUSF
type UEAuthenticationResponse struct {
	AuthType      string                 `json:"authType"`
	AuthCtxID     string                 `json:"authCtxId"`
	Var5gAuthData *Var5gAuthData         `json:"_5gAuthData,omitempty"`
	Links         map[string]interface{} `json:"_links"`
}

// Var5gAuthData represents 5G authentication data
type Var5gAuthData struct {
	RAND string `json:"rand"`
	AUTN string `json:"autn"`
}

// AuthConfirmationRequest represents authentication confirmation
type AuthConfirmationRequest struct {
	RES string `json:"resStar"`
}

// AuthConfirmationResponse represents confirmation response
type AuthConfirmationResponse struct {
	AuthResult string `json:"authResult"` // "AUTHENTICATION_SUCCESS" or "AUTHENTICATION_FAILURE"
	SUPI       string `json:"supi,omitempty"`
	KSEAF      string `json:"kseaf,omitempty"`
}

// InitiateAuthentication initiates UE authentication with AUSF
func (c *AUSFClient) InitiateAuthentication(ctx context.Context, req *UEAuthenticationRequest) (*UEAuthenticationResponse, error) {
	url := fmt.Sprintf("%s/nausf-auth/v1/ue-authentications", c.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	c.logger.Debug("Initiating authentication with AUSF",
		zap.String("supi", req.SUPI),
		zap.String("url", url),
	)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AUSF returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result UEAuthenticationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Debug("Authentication initiated with AUSF",
		zap.String("supi", req.SUPI),
		zap.String("auth_ctx_id", result.AuthCtxID),
	)

	return &result, nil
}

// ConfirmAuthentication confirms authentication with AUSF
func (c *AUSFClient) ConfirmAuthentication(ctx context.Context, authCtxID string, resStar string) (*AuthConfirmationResponse, error) {
	url := fmt.Sprintf("%s/nausf-auth/v1/ue-authentications/%s/5g-aka-confirmation", c.baseURL, authCtxID)

	req := &AuthConfirmationRequest{
		RES: resStar,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AUSF returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result AuthConfirmationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Debug("Authentication confirmed with AUSF",
		zap.String("auth_ctx_id", authCtxID),
		zap.String("result", result.AuthResult),
	)

	return &result, nil
}
