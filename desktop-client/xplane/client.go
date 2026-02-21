package xplane

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Position holds the current flight position data
type Position struct {
	Latitude    float64
	Longitude   float64
	AltitudeAGL float64 // meters
	Groundspeed float64 // m/s
	Heading     float64 // magnetic heading
	TailNumber  string
	Timestamp   time.Time
}

// IsValid returns true if we have received position data
func (p *Position) IsValid() bool {
	return p.Latitude != 0 || p.Longitude != 0
}

// Client handles WebSocket communication with X-Plane
type Client struct {
	port         int
	conn         *websocket.Conn
	datarefMap   DatarefMap
	reverseMap   map[int64]string
	position     Position
	positionMu   sync.RWMutex
	connected    bool
	connectedMu  sync.RWMutex
	onConnect    func()
	onDisconnect func()
	stopCh       chan struct{}
	doneCh       chan struct{} // signals when connection is lost
}

// WebSocket message types
type wsMessage struct {
	ReqID  int         `json:"req_id"`
	Type   string      `json:"type"`
	Params interface{} `json:"params,omitempty"`
}

type subscribeParams struct {
	Datarefs []datarefSub `json:"datarefs"`
}

type datarefSub struct {
	ID int64 `json:"id"`
}

type wsResponse struct {
	ReqID   int                    `json:"req_id"`
	Type    string                 `json:"type"`
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// NewClient creates a new X-Plane client
func NewClient(port int) *Client {
	return &Client{
		port:   port,
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}
}

// Done returns a channel that is closed when the connection is lost
func (c *Client) Done() <-chan struct{} {
	return c.doneCh
}

// SetCallbacks sets connection state callbacks
func (c *Client) SetCallbacks(onConnect, onDisconnect func()) {
	c.onConnect = onConnect
	c.onDisconnect = onDisconnect
}

// Connect resolves dataref IDs and establishes WebSocket connection
func (c *Client) Connect() error {
	// Step 1: Resolve dataref names to session IDs via REST API
	datarefMap, err := ResolveDatarefIDs(c.port, AllDatarefs)
	if err != nil {
		return fmt.Errorf("failed to resolve datarefs: %w", err)
	}
	c.datarefMap = datarefMap
	c.reverseMap = datarefMap.ReverseMap()

	// Step 2: Connect to WebSocket
	wsURL := fmt.Sprintf("ws://localhost:%d/api/v3", c.port)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("WebSocket connection failed: %w", err)
	}
	c.conn = conn

	// Step 3: Subscribe to datarefs using numeric IDs
	if err := c.subscribe(); err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	c.setConnected(true)
	if c.onConnect != nil {
		c.onConnect()
	}

	// Start reading messages
	go c.readLoop()

	return nil
}

// subscribe sends subscription request for all datarefs
func (c *Client) subscribe() error {
	var subs []datarefSub
	for _, id := range c.datarefMap {
		subs = append(subs, datarefSub{ID: id})
	}

	msg := wsMessage{
		ReqID: 1,
		Type:  "dataref_subscribe_values",
		Params: subscribeParams{
			Datarefs: subs,
		},
	}

	return c.conn.WriteJSON(msg)
}

// readLoop continuously reads WebSocket messages and updates position
func (c *Client) readLoop() {
	defer func() {
		c.setConnected(false)
		close(c.doneCh)
		if c.onDisconnect != nil {
			c.onDisconnect()
		}
	}()

	for {
		select {
		case <-c.stopCh:
			return
		default:
		}

		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			return
		}

		// Debug: log first few messages
		log.Printf("WS message: %s", string(message[:min(len(message), 500)]))

		var resp wsResponse
		if err := json.Unmarshal(message, &resp); err != nil {
			log.Printf("JSON unmarshal error: %v", err)
			continue
		}

		// Process dataref values
		if resp.Data != nil {
			c.updatePosition(resp.Data)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// updatePosition updates the current position from WebSocket data
func (c *Client) updatePosition(data map[string]interface{}) {
	c.positionMu.Lock()
	defer c.positionMu.Unlock()

	for idStr, value := range data {
		// Parse the string ID back to int64
		var id int64
		fmt.Sscanf(idStr, "%d", &id)

		name, ok := c.reverseMap[id]
		if !ok {
			continue
		}

		// Values come directly (not wrapped in {"value": ...})
		switch name {
		case DatarefLatitude:
			if v, ok := value.(float64); ok {
				c.position.Latitude = v
			}
		case DatarefLongitude:
			if v, ok := value.(float64); ok {
				c.position.Longitude = v
			}
		case DatarefAltitudeAGL:
			if v, ok := value.(float64); ok {
				c.position.AltitudeAGL = v
			}
		case DatarefGroundspeed:
			if v, ok := value.(float64); ok {
				c.position.Groundspeed = v
			}
		case DatarefHeading:
			if v, ok := value.(float64); ok {
				c.position.Heading = v
			}
		case DatarefTailNum:
			c.position.TailNumber = DecodeTailNumber(value)
		}
	}
	c.position.Timestamp = time.Now()
}

// GetPosition returns the current position (thread-safe)
func (c *Client) GetPosition() Position {
	c.positionMu.RLock()
	defer c.positionMu.RUnlock()
	return c.position
}

// IsConnected returns true if connected to X-Plane
func (c *Client) IsConnected() bool {
	c.connectedMu.RLock()
	defer c.connectedMu.RUnlock()
	return c.connected
}

func (c *Client) setConnected(connected bool) {
	c.connectedMu.Lock()
	c.connected = connected
	c.connectedMu.Unlock()
}

// Disconnect closes the WebSocket connection
func (c *Client) Disconnect() {
	close(c.stopCh)
	if c.conn != nil {
		c.conn.Close()
	}
	c.setConnected(false)
}
