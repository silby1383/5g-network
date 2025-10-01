package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/your-org/5g-network/nf/smf/internal/config"
	"go.uber.org/zap"
)

// NRFClient handles communication with NRF
type NRFClient struct {
	config       *config.Config
	httpClient   *http.Client
	logger       *zap.Logger
	nfInstanceID string
}

// NewNRFClient creates a new NRF client
func NewNRFClient(cfg *config.Config, logger *zap.Logger) *NRFClient {
	return &NRFClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger:       logger,
		nfInstanceID: generateNFInstanceID("smf"),
	}
}

// NFProfile represents SMF's NF profile for registration
type NFProfile struct {
	NFInstanceID   string      `json:"nfInstanceId"`
	NFType         string      `json:"nfType"`
	NFStatus       string      `json:"nfStatus"`
	PLMNID         PLMNID      `json:"plmnId"`
	SNSSAI         []SNSSAI    `json:"sNssai"`
	IPv4Addresses  []string    `json:"ipv4Addresses"`
	NFServices     []NFService `json:"nfServices"`
	HeartBeatTimer int         `json:"heartBeatTimer"`
}

// PLMNID represents PLMN identifier
type PLMNID struct {
	MCC string `json:"mcc"`
	MNC string `json:"mnc"`
}

// SNSSAI represents S-NSSAI
type SNSSAI struct {
	SST int    `json:"sst"`
	SD  string `json:"sd"`
}

// NFService represents NF service
type NFService struct {
	ServiceInstanceID string             `json:"serviceInstanceId"`
	ServiceName       string             `json:"serviceName"`
	Versions          []NFServiceVersion `json:"versions"`
	Scheme            string             `json:"scheme"`
	NfServiceStatus   string             `json:"nfServiceStatus"`
	IPv4EndPoints     []string           `json:"ipv4EndPoints"`
}

// NFServiceVersion represents NF service version
type NFServiceVersion struct {
	APIVersionInURI string `json:"apiVersionInUri"`
	APIFullVersion  string `json:"apiFullVersion"`
}

// Register registers SMF with NRF
func (c *NRFClient) Register() error {
	url := fmt.Sprintf("%s/nnrf-nfm/v1/nf-instances/%s", c.config.NRF.URL, c.nfInstanceID)

	// Build SNSSAI list
	var snssai []SNSSAI
	for _, s := range c.config.SMF.SupportedSNSSAI {
		snssai = append(snssai, SNSSAI{
			SST: s.SST,
			SD:  s.SD,
		})
	}

	profile := NFProfile{
		NFInstanceID: c.nfInstanceID,
		NFType:       "SMF",
		NFStatus:     "REGISTERED",
		PLMNID: PLMNID{
			MCC: c.config.SMF.PLMN.MCC,
			MNC: c.config.SMF.PLMN.MNC,
		},
		SNSSAI:         snssai,
		IPv4Addresses:  []string{c.config.SBI.IPv4},
		HeartBeatTimer: 30,
		NFServices: []NFService{
			{
				ServiceInstanceID: "nsmf-pdusession",
				ServiceName:       "nsmf-pdusession",
				Versions: []NFServiceVersion{
					{
						APIVersionInURI: "v1",
						APIFullVersion:  "1.0.0",
					},
				},
				Scheme:          c.config.SBI.Scheme,
				NfServiceStatus: "REGISTERED",
				IPv4EndPoints:   []string{fmt.Sprintf("%s:%d", c.config.SBI.IPv4, c.config.SBI.Port)},
			},
		},
	}

	body, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal NF profile: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	c.logger.Info("Registering SMF with NRF",
		zap.String("nrf_url", c.config.NRF.URL),
		zap.String("nf_instance_id", c.nfInstanceID),
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send registration request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("NRF registration failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	c.logger.Info("SMF registered successfully with NRF")
	return nil
}

// SendHeartbeat sends heartbeat to NRF
func (c *NRFClient) SendHeartbeat() error {
	url := fmt.Sprintf("%s/nnrf-nfm/v1/nf-instances/%s/heartbeat", c.config.NRF.URL, c.nfInstanceID)

	req, err := http.NewRequest(http.MethodPatch, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create heartbeat request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send heartbeat: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("heartbeat failed with status %d", resp.StatusCode)
	}

	c.logger.Debug("Heartbeat sent to NRF")
	return nil
}

// Deregister deregisters SMF from NRF
func (c *NRFClient) Deregister() error {
	url := fmt.Sprintf("%s/nnrf-nfm/v1/nf-instances/%s", c.config.NRF.URL, c.nfInstanceID)

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create deregistration request: %w", err)
	}

	c.logger.Info("Deregistering SMF from NRF")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send deregistration request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("deregistration failed with status %d", resp.StatusCode)
	}

	c.logger.Info("SMF deregistered successfully from NRF")
	return nil
}

// generateNFInstanceID generates a unique NF instance ID
func generateNFInstanceID(nfType string) string {
	return fmt.Sprintf("%s-%d", nfType, time.Now().UnixNano())
}
