package xplane

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Dataref names we need to subscribe to
const (
	DatarefLatitude    = "sim/flightmodel/position/latitude"
	DatarefLongitude   = "sim/flightmodel/position/longitude"
	DatarefAltitudeAGL = "sim/flightmodel/position/y_agl"
	DatarefGroundspeed = "sim/flightmodel/position/groundspeed"
	DatarefHeading     = "sim/flightmodel/position/mag_psi"
	DatarefTailNum     = "sim/aircraft/view/acf_tailnum"
)

// AllDatarefs is the list of all datarefs we need
var AllDatarefs = []string{
	DatarefLatitude,
	DatarefLongitude,
	DatarefAltitudeAGL,
	DatarefGroundspeed,
	DatarefHeading,
	DatarefTailNum,
}

// DatarefInfo holds metadata about a dataref
type DatarefInfo struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	ValueType string `json:"value_type"`
}

// DatarefResponse represents the X-Plane REST API response for dataref queries
type DatarefResponse struct {
	Data []DatarefInfo `json:"data"`
}

// DatarefMap maps dataref names to their session-specific IDs
type DatarefMap map[string]int64

// ReverseMap returns a map from ID to dataref name
func (m DatarefMap) ReverseMap() map[int64]string {
	reverse := make(map[int64]string)
	for name, id := range m {
		reverse[id] = name
	}
	return reverse
}

// ResolveDatarefIDs queries the X-Plane REST API to get session-specific IDs for datarefs
func ResolveDatarefIDs(port int, datarefs []string) (DatarefMap, error) {
	result := make(DatarefMap)
	client := &http.Client{Timeout: 5 * time.Second}

	for _, name := range datarefs {
		id, err := resolveDataref(client, port, name)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve %s: %w", name, err)
		}
		result[name] = id
	}

	return result, nil
}

// resolveDataref queries X-Plane for a single dataref's ID
func resolveDataref(client *http.Client, port int, name string) (int64, error) {
	// Use raw brackets - X-Plane may not handle URL-encoded brackets
	apiURL := fmt.Sprintf("http://localhost:%d/api/v3/datarefs?filter[name]=%s", port, url.PathEscape(name))

	log.Printf("Requesting: %s", apiURL)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		preview := string(body)
		if len(preview) > 200 {
			preview = preview[:200]
		}
		log.Printf("Response body: %s", preview)
		return 0, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var datarefResp DatarefResponse
	if err := json.NewDecoder(resp.Body).Decode(&datarefResp); err != nil {
		return 0, err
	}

	if len(datarefResp.Data) == 0 {
		return 0, fmt.Errorf("dataref not found")
	}

	return datarefResp.Data[0].ID, nil
}

// DecodeTailNumber decodes the tail number from X-Plane's format
// X-Plane returns byte arrays as base64 strings or int arrays
func DecodeTailNumber(value interface{}) string {
	switch v := value.(type) {
	case string:
		// Try base64 decode
		decoded, err := base64.StdEncoding.DecodeString(v)
		if err == nil {
			return cleanTailNumber(string(decoded))
		}
		// Already a plain string
		return cleanTailNumber(v)

	case []interface{}:
		// Int array of byte values
		var bytes []byte
		for _, b := range v {
			if num, ok := b.(float64); ok {
				if num == 0 {
					break // Null terminator
				}
				bytes = append(bytes, byte(num))
			}
		}
		return cleanTailNumber(string(bytes))

	default:
		return "UNKNOWN"
	}
}

// cleanTailNumber removes null bytes and trims whitespace
func cleanTailNumber(s string) string {
	// Remove null bytes
	s = strings.ReplaceAll(s, "\x00", "")
	// Trim whitespace
	s = strings.TrimSpace(s)
	if s == "" {
		return "UNKNOWN"
	}
	return s
}
