package mikrotik

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Client represents a MikroTik REST API client
type Client struct {
	Host     string
	Username string
	Password string
	Client   *http.Client
}

// NewClient creates a new MikroTik REST API client
func NewClient(host, username, password string) *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &Client{
		Host:     host,
		Username: username,
		Password: password,
		Client: &http.Client{
			Transport: tr,
			Timeout:   30 * time.Second,
		},
	}
}

// REST API request/response types
type apiRequest struct {
	Path    string      `json:"path"`
	Content interface{} `json:"content,omitempty"`
}

type apiResponse struct {
	After []interface{} `json:"after,omitempty"`
	From  []interface{} `json:"from,omitempty"`
	Tags  []interface{} `json:"tags,omitempty"`
	Error *apiError     `json:"error,omitempty"`
}

type apiError struct {
	Category int    `json:"category"`
	Detail   string `json:"detail"`
	Message  string `json:"message"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"ret"`
}

// doRequest performs an authenticated REST API request
func (c *Client) doRequest(method, path string, payload interface{}) (*apiResponse, error) {
	var body []byte
	var err error

	url := fmt.Sprintf("https://%s/rest%s", c.Host, path)

	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Try with basic auth first (works for most MikroTik versions)
	req.SetBasicAuth(c.Username, c.Password)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var result apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		// Some endpoints return empty response on success
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return &apiResponse{}, nil
		}
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode >= 400 {
		if result.Error != nil {
			return nil, fmt.Errorf("API error: %s - %s", result.Error.Message, result.Error.Detail)
		}
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	return &result, nil
}

// Reboot sends a reboot command to the MikroTik device
func (c *Client) Reboot() error {
	zap.S().Infow("Sending reboot command to MikroTik",
		"host", c.Host,
		"namespace", "mikrotik")

	// MikroTik REST API for system reboot
	// POST /rest/system/reboot with content: {".id": "*0"}
	payload := map[string]interface{}{
		".id": "*0",
	}

	resp, err := c.doRequest("POST", "/system/reboot", payload)
	if err != nil {
		return fmt.Errorf("reboot failed: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("reboot error: %s", resp.Error.Message)
	}

	zap.S().Infow("Reboot command sent successfully",
		"host", c.Host,
		"namespace", "mikrotik")

	return nil
}

// ExecuteCommand executes a raw MikroTik command
func (c *Client) ExecuteCommand(path string, content map[string]interface{}) error {
	resp, err := c.doRequest("POST", path, content)
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return fmt.Errorf("command error: %s", resp.Error.Message)
	}

	return nil
}

// GetSystemInfo retrieves system information from the device
func (c *Client) GetSystemInfo() (map[string]interface{}, error) {
	resp, err := c.doRequest("GET", "/system/resource/print", nil)
	if err != nil {
		return nil, err
	}

	// Parse the response - MikroTik REST API returns array
	if len(resp.After) > 0 {
		if data, ok := resp.After[0].(map[string]interface{}); ok {
			return data, nil
		}
	}

	return nil, fmt.Errorf("no data in response")
}

// HealthCheck verifies connectivity to the MikroTik device
func (c *Client) HealthCheck() error {
	_, err := c.GetSystemInfo()
	return err
}
