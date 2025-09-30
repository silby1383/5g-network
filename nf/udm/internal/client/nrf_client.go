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

// NRFClient handles communication with NRF
type NRFClient struct {
	baseURL string
	client  *http.Client
	logger  *zap.Logger
}

// NewNRFClient creates a new NRF client
func NewNRFClient(baseURL string, logger *zap.Logger) *NRFClient {
	return &NRFClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// NFProfile represents an NF profile for registration
type NFProfile struct {
	NFInstanceID string   `json:"nfInstanceId"`
	NFType       string   `json:"nfType"`
	NFStatus     string   `json:"nfStatus"`
	PLMNID       PLMNID   `json:"plmnId"`
	SNSSAI       []SNSSAI `json:"sNssais,omitempty"`
	IPv4Address  string   `json:"ipv4Addresses,omitempty"`
	Capacity     int      `json:"capacity,omitempty"`
	Priority     int      `json:"priority,omitempty"`
	UDMInfo      *UDMInfo `json:"udmInfo,omitempty"`
}

// PLMNID represents PLMN identifier
type PLMNID struct {
	MCC string `json:"mcc"`
	MNC string `json:"mnc"`
}

// UDMInfo contains UDM-specific information
type UDMInfo struct {
	GroupID               string   `json:"groupId,omitempty"`
	SupiRanges            []string `json:"supiRanges,omitempty"`
	GPSIRanges            []string `json:"gpsiRanges,omitempty"`
	ExternalGroupID       []string `json:"externalGroupIds,omitempty"`
	RoutingIndicators     []string `json:"routingIndicators,omitempty"`
	InternalGroupID       []string `json:"internalGroupIds,omitempty"`
}

// Register registers UDM with NRF
func (c *NRFClient) Register(ctx context.Context, profile *NFProfile) error {
	url := fmt.Sprintf("%s/nnrf-nfm/v1/nf-instances/%s", c.baseURL, profile.NFInstanceID)

	body, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("NRF returned status %d: %s", resp.StatusCode, string(respBody))
	}

	c.logger.Info("Registered with NRF", zap.String("nf_instance_id", profile.NFInstanceID))
	return nil
}

// Deregister removes UDM registration from NRF
func (c *NRFClient) Deregister(ctx context.Context, nfInstanceID string) error {
	url := fmt.Sprintf("%s/nnrf-nfm/v1/nf-instances/%s", c.baseURL, nfInstanceID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("NRF returned status %d: %s", resp.StatusCode, string(respBody))
	}

	c.logger.Info("Deregistered from NRF", zap.String("nf_instance_id", nfInstanceID))
	return nil
}

// Heartbeat sends heartbeat to NRF
func (c *NRFClient) Heartbeat(ctx context.Context, nfInstanceID string) error {
	url := fmt.Sprintf("%s/nnrf-nfm/v1/nf-instances/%s/heartbeat", c.baseURL, nfInstanceID)

	req, err := http.NewRequestWithContext(ctx, "PATCH", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("NRF returned status %d: %s", resp.StatusCode, string(respBody))
	}

	c.logger.Debug("Heartbeat sent to NRF", zap.String("nf_instance_id", nfInstanceID))
	return nil
}
