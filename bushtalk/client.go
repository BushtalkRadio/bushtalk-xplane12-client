package bushtalk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	ClientName    = "XPLANE-12"
	ClientVersion = "1.0.0"
)

// Client handles communication with the Bushtalk Radio API
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// setHeaders adds common headers to all requests
func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-BTR-CLIENT-NAME", ClientName)
	req.Header.Set("X-BTR-CLIENT-VERSION", ClientVersion)
}

// AuthResponse represents the authentication response from the API
type AuthResponse struct {
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"userId"`
	Username     string `json:"username"`
	ExpiresIn    int    `json:"expires_in"`
}

// TrackPayload represents flight position data sent to the API
type TrackPayload struct {
	Latitude       float64 `json:"PLANE_LATITUDE"`
	Longitude      float64 `json:"PLANE_LONGITUDE"`
	AltitudeAGL    float64 `json:"ALTITUDE_ABOVE_GROUND"`
	GroundVelocity float64 `json:"GROUND_VELOCITY"`
	Heading        float64 `json:"MAGNETIC_COMPASS"`
	TailNumber     string  `json:"ATC_ID"`
	OnGround       bool    `json:"SIM_ON_GROUND"`
}

// NewClient creates a new Bushtalk API client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SetToken sets the authentication token for API requests
func (c *Client) SetToken(token string) {
	c.token = token
}

// GetToken returns the current authentication token
func (c *Client) GetToken() string {
	return c.token
}

// Authenticate logs in with username/password and returns auth response
func (c *Client) Authenticate(username, password string) (*AuthResponse, error) {
	payload := map[string]string{
		"username": username,
		"password": password,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal auth request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/authenticate", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create auth request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("authentication failed: status %d", resp.StatusCode)
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("failed to decode auth response: %w", err)
	}

	c.token = authResp.IDToken
	return &authResp, nil
}

// SendPosition sends flight position data to the tracking API
func (c *Client) SendPosition(payload *TrackPayload) error {
	if c.token == "" {
		return fmt.Errorf("not authenticated")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal position: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/track", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create track request: %w", err)
	}
	c.setHeaders(req)
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("track request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("track request failed: status %d", resp.StatusCode)
	}

	return nil
}
