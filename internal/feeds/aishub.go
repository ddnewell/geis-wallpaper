package feeds

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// AISHubConfig configures the keyed AISHub poll. Used server-side only so the
// membership username never reaches the client. AISHub enforces a >=1 minute
// poll interval, which suits a cron-driven refresh.
type AISHubConfig struct {
	Username       string
	LatMin, LatMax float64
	LonMin, LonMax float64
}

func (c AISHubConfig) hasBBox() bool {
	return c.LatMin != 0 || c.LatMax != 0 || c.LonMin != 0 || c.LonMax != 0
}

// FetchShips polls AISHub for vessel positions.
func FetchShips(ctx context.Context, cfg AISHubConfig) ([]Ship, error) {
	if cfg.Username == "" {
		return nil, fmt.Errorf("aishub username not configured")
	}
	q := url.Values{}
	q.Set("username", cfg.Username)
	q.Set("format", "1") // human-readable decimal lat/lon
	q.Set("output", "json")
	q.Set("compress", "0")
	if cfg.hasBBox() {
		q.Set("latmin", strconv.FormatFloat(cfg.LatMin, 'f', 4, 64))
		q.Set("latmax", strconv.FormatFloat(cfg.LatMax, 'f', 4, 64))
		q.Set("lonmin", strconv.FormatFloat(cfg.LonMin, 'f', 4, 64))
		q.Set("lonmax", strconv.FormatFloat(cfg.LonMax, 'f', 4, 64))
	}
	b, err := GetBytes(ctx, "https://data.aishub.net/ws.php?"+q.Encode())
	if err != nil {
		return nil, err
	}

	// Response is [ {metadata}, [ {vessel}, ... ] ].
	var raw []json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil || len(raw) == 0 {
		return nil, fmt.Errorf("unexpected AISHub response")
	}
	var meta struct {
		Error   bool   `json:"ERROR"`
		Message string `json:"ERROR_MESSAGE"`
	}
	if err := json.Unmarshal(raw[0], &meta); err != nil {
		return nil, fmt.Errorf("decode AISHub metadata: %w", err)
	}
	if meta.Error {
		return nil, fmt.Errorf("AISHub error: %s", meta.Message)
	}
	if len(raw) < 2 {
		return []Ship{}, nil
	}
	var vessels []struct {
		MMSI json.Number `json:"MMSI"`
		Lat  float64     `json:"LATITUDE"`
		Lon  float64     `json:"LONGITUDE"`
		Name string      `json:"NAME"`
	}
	if err := json.Unmarshal(raw[1], &vessels); err != nil {
		return nil, fmt.Errorf("decode AISHub vessels: %w", err)
	}
	out := make([]Ship, 0, len(vessels))
	for _, v := range vessels {
		out = append(out, Ship{Lat: v.Lat, Lon: v.Lon, Name: v.Name, MMSI: v.MMSI.String()})
	}
	return out, nil
}
