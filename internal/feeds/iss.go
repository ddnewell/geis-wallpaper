package feeds

import (
	"context"
	"time"
)

// issURL is the keyless wheretheiss.at current-position endpoint for the ISS
// (NORAD 25544). open-notify was found unreliable; this one is stable.
const issURL = "https://api.wheretheiss.at/v1/satellites/25544"

type issResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp int64   `json:"timestamp"` // sec epoch
}

// FetchISS returns the current ISS position.
func FetchISS(ctx context.Context) (*ISS, error) {
	var r issResponse
	if err := GetJSON(ctx, issURL, &r); err != nil {
		return nil, err
	}
	return &ISS{Lat: r.Latitude, Lon: r.Longitude, Time: time.Unix(r.Timestamp, 0)}, nil
}
