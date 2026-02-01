// Package media provides Telegram location operations.
package media

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gotd/td/tg"
	"agent-telegram/telegram/types"
)

// GeoCoder handles geocoding requests.
type GeoCoder struct {
	client *http.Client
}

// NewGeoCoder creates a new GeoCoder.
func NewGeoCoder() *GeoCoder {
	return &GeoCoder{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// GeocodeResult represents a geocoding result.
type GeocodeResult struct {
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	DisplayName string  `json:"display_name"` //nolint:tagliatelle // External API format
}

// Geocode converts a city name to coordinates using Nominatim API.
func (g *GeoCoder) Geocode(city string) (*GeocodeResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build URL
	u, err := url.Parse("https://nominatim.openstreetmap.org/search")
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("q", city)
	params.Set("format", "json")
	params.Set("limit", "1")
	u.RawQuery = params.Encode()

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	// Make request
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("geocode request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geocode failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var results []struct {
		Lat         string `json:"lat"`
		Lon         string `json:"lon"`
		DisplayName string `json:"display_name"` //nolint:tagliatelle // External API format
	}

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to parse geocode response: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("location not found: %s", city)
	}

	// Parse coordinates
	lat, err := strconv.ParseFloat(results[0].Lat, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid latitude: %w", err)
	}
	lon, err := strconv.ParseFloat(results[0].Lon, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid longitude: %w", err)
	}

	return &GeocodeResult{
		Lat:         lat,
		Lon:         lon,
		DisplayName: results[0].DisplayName,
	}, nil
}

// SendLocation sends a location to a peer.
func (c *Client) SendLocation(ctx context.Context, params types.SendLocationParams) (*types.SendLocationResult, error) {
	if c.API == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	// Clean peer (remove @ prefix)
	peer := strings.TrimPrefix(params.Peer, "@")

	// Resolve username to get input peer
	inputPeer, err := c.ResolvePeer(ctx, peer)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve peer @%s: %w", peer, err)
	}

	// Create geo point media
	geoPoint := &tg.InputMediaGeoPoint{
		GeoPoint: &tg.InputGeoPoint{
			Lat:  params.Latitude,
			Long: params.Longitude,
		},
	}

	// Send location using MessagesSendMedia
	result, err := c.API.MessagesSendMedia(ctx, &tg.MessagesSendMediaRequest{
		Peer:     inputPeer,
		Media:    geoPoint,
		RandomID: time.Now().UnixNano(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send location: %w", err)
	}

	// Extract message ID from response
	msgID := extractMessageID(result)

	return &types.SendLocationResult{
		ID:        msgID,
		Date:      time.Now().Unix(),
		Peer:      peer,
		Latitude:  params.Latitude,
		Longitude: params.Longitude,
	}, nil
}
