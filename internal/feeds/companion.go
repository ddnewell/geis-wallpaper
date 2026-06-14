package feeds

import (
	"context"
	"strings"
)

// CompanionPayload fetches the aggregated payload from the companion server.
func CompanionPayload(ctx context.Context, baseURL string) (*Payload, error) {
	var p Payload
	if err := GetJSON(ctx, strings.TrimRight(baseURL, "/")+"/payload.json", &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// CompanionSatellite fetches the pre-resized cloud image from the companion.
func CompanionSatellite(ctx context.Context, baseURL string) ([]byte, error) {
	return GetBytes(ctx, strings.TrimRight(baseURL, "/")+"/satellite.png")
}
