package feeds

import (
	"context"
	"time"
)

// usgsAllDay is the keyless USGS GeoJSON feed of all earthquakes in the past day.
const usgsAllDay = "https://earthquake.usgs.gov/earthquakes/feed/v1.0/summary/all_day.geojson"

type usgsCollection struct {
	Features []struct {
		Properties struct {
			Mag   *float64 `json:"mag"`
			Place string   `json:"place"`
			Time  int64    `json:"time"` // ms epoch
		} `json:"properties"`
		Geometry struct {
			Coordinates []float64 `json:"coordinates"` // [lon, lat, depthKm]
		} `json:"geometry"`
	} `json:"features"`
}

// FetchQuakes returns earthquakes at or above minMag from the USGS feed.
func FetchQuakes(ctx context.Context, minMag float64) ([]Quake, error) {
	var fc usgsCollection
	if err := GetJSON(ctx, usgsAllDay, &fc); err != nil {
		return nil, err
	}
	out := make([]Quake, 0, len(fc.Features))
	for _, f := range fc.Features {
		if f.Properties.Mag == nil || *f.Properties.Mag < minMag {
			continue
		}
		c := f.Geometry.Coordinates
		if len(c) < 2 {
			continue
		}
		depth := 0.0
		if len(c) >= 3 {
			depth = c[2]
		}
		out = append(out, Quake{
			Mag:     *f.Properties.Mag,
			Lon:     c[0],
			Lat:     c[1],
			DepthKm: depth,
			Place:   f.Properties.Place,
			Time:    time.UnixMilli(f.Properties.Time),
		})
	}
	return out, nil
}
