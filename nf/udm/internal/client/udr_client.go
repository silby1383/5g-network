package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// UDRClient handles communication with UDR
type UDRClient struct {
	baseURL string
	client  *http.Client
	logger  *zap.Logger
}

// NewUDRClient creates a new UDR client
func NewUDRClient(baseURL string, timeout time.Duration, logger *zap.Logger) *UDRClient {
	return &UDRClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: timeout,
		},
		logger: logger,
	}
}

// SubscriberData represents subscriber information from UDR
type SubscriberData struct {
	SUPI                     string                 `json:"supi"`
	SUPIType                 string                 `json:"supiType"`
	PLMNIDmcc                string                 `json:"plmnId.mcc"`
	PLMNIDmnc                string                 `json:"plmnId.mnc"`
	SubscriberStatus         string                 `json:"subscriberStatus,omitempty"`
	MSISDN                   string                 `json:"msisdn,omitempty"`
	SubscribedUeAmbrUplink   uint64                 `json:"subscribedUeAmbr.uplink,string"`
	SubscribedUeAmbrDownlink uint64                 `json:"subscribedUeAmbr.downlink,string"`
	NSSAI                    []SNSSAI               `json:"nssai,omitempty"`
	DNNConfigurations        map[string]interface{} `json:"dnnConfigurations,omitempty"`
	RoamingAllowed           bool                   `json:"roamingAllowed"`
}

// SNSSAI represents Single Network Slice Selection Assistance Information
type SNSSAI struct {
	SST int    `json:"sst"`
	SD  string `json:"sd,omitempty"`
}

// AuthenticationSubscription represents authentication data from UDR
type AuthenticationSubscription struct {
	SUPI                          string `json:"supi"`
	AuthenticationMethod          string `json:"authenticationMethod"`
	PermanentKey                  string `json:"permanentKey"`
	PermanentKeyID                uint8  `json:"permanentKeyId"`
	EncAlgorithm                  string `json:"encAlgorithm"`
	EncOPC                        string `json:"encOpc"`
	EncOP                         string `json:"encTopcKey,omitempty"`
	SQN                           uint64 `json:"sequenceNumber,string"`
	SQNScheme                     string `json:"sqnScheme"`
	AuthenticationManagementField string `json:"authenticationManagementField"`
}

// SessionManagementSubscriptionData represents SM subscription data
type SessionManagementSubscriptionData struct {
	SUPI                  string   `json:"supi"`
	DNN                   string   `json:"dnn"`
	SessionAmbrUplink     uint64   `json:"sessionAmbrUplink"`
	SessionAmbrDownlink   uint64   `json:"sessionAmbrDownlink"`
	Default5QI            uint8    `json:"default5qi"`
	ARPPriorityLevel      uint8    `json:"arpPriorityLevel"`
	SSCModes              []uint8  `json:"sscModes"`
	DefaultSSCMode        uint8    `json:"defaultSscMode"`
	PDUSessionTypes       []string `json:"pduSessionTypes"`
	DefaultPDUSessionType string   `json:"defaultPduSessionType"`
}

// GetSubscriberData retrieves subscriber data from UDR
func (c *UDRClient) GetSubscriberData(ctx context.Context, supi string) (*SubscriberData, error) {
	url := fmt.Sprintf("%s/admin/subscribers/%s", c.baseURL, supi)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("UDR returned status %d: %s", resp.StatusCode, string(body))
	}

	var data SubscriberData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Debug("Retrieved subscriber data from UDR", zap.String("supi", supi))
	return &data, nil
}

// GetAuthenticationSubscription retrieves authentication subscription from UDR
func (c *UDRClient) GetAuthenticationSubscription(ctx context.Context, supi string) (*AuthenticationSubscription, error) {
	url := fmt.Sprintf("%s/nudr-dr/v1/subscription-data/%s/authentication-data/authentication-subscription", c.baseURL, supi)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("UDR returned status %d: %s", resp.StatusCode, string(body))
	}

	var data AuthenticationSubscription
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Debug("Retrieved authentication subscription from UDR", zap.String("supi", supi))
	return &data, nil
}

// IncrementSQN increments the sequence number in UDR
func (c *UDRClient) IncrementSQN(ctx context.Context, supi string) (uint64, error) {
	url := fmt.Sprintf("%s/nudr-dr/v1/subscription-data/%s/authentication-data/authentication-subscription/sqn", c.baseURL, supi)

	req, err := http.NewRequestWithContext(ctx, "PATCH", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("UDR returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		SQN uint64 `json:"sqn"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Debug("Incremented SQN in UDR", zap.String("supi", supi), zap.Uint64("new_sqn", result.SQN))
	return result.SQN, nil
}

// GetSessionManagementData retrieves session management subscription data
func (c *UDRClient) GetSessionManagementData(ctx context.Context, supi, dnn string) (*SessionManagementSubscriptionData, error) {
	url := fmt.Sprintf("%s/nudr-dr/v1/subscription-data/%s/provisioned-data/sm-data?dnn=%s", c.baseURL, supi, dnn)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("UDR returned status %d: %s", resp.StatusCode, string(body))
	}

	var data SessionManagementSubscriptionData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.Debug("Retrieved SM data from UDR", zap.String("supi", supi), zap.String("dnn", dnn))
	return &data, nil
}
