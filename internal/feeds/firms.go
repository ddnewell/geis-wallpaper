package feeds

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// FIRMSConfig configures the keyed NASA FIRMS active-fire query. Used
// server-side only so the MAP_KEY stays off the client.
type FIRMSConfig struct {
	MapKey   string
	Source   string  // e.g. "VIIRS_SNPP_NRT"
	Area     string  // "world" or "west,south,east,north"
	DayRange int     // 1..10
	MinFRP   float64 // drop detections below this fire radiative power (MW); 0 = keep all
}

// FetchFires queries the FIRMS area CSV API and returns active-fire detections.
func FetchFires(ctx context.Context, cfg FIRMSConfig) ([]Fire, error) {
	if cfg.MapKey == "" {
		return nil, fmt.Errorf("FIRMS map key not configured")
	}
	source := cfg.Source
	if source == "" {
		source = "VIIRS_SNPP_NRT"
	}
	area := cfg.Area
	if area == "" {
		area = "world"
	}
	day := cfg.DayRange
	if day < 1 {
		day = 1
	}
	url := fmt.Sprintf("https://firms.modaps.eosdis.nasa.gov/api/area/csv/%s/%s/%s/%d",
		cfg.MapKey, source, area, day)
	b, err := GetBytes(ctx, url)
	if err != nil {
		return nil, err
	}
	if bytes.HasPrefix(bytes.TrimSpace(b), []byte("Invalid")) {
		return nil, fmt.Errorf("FIRMS rejected request (check map key / params)")
	}

	r := csv.NewReader(bytes.NewReader(b))
	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("FIRMS CSV header: %w", err)
	}
	col := map[string]int{}
	for i, h := range header {
		col[strings.TrimSpace(strings.ToLower(h))] = i
	}
	latI, okLat := col["latitude"]
	lonI, okLon := col["longitude"]
	if !okLat || !okLon {
		return nil, fmt.Errorf("FIRMS CSV missing latitude/longitude columns")
	}
	confI, hasConf := col["confidence"]
	frpI, hasFRP := col["frp"]

	out := []Fire{}
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		lat, e1 := strconv.ParseFloat(field(rec, latI), 64)
		lon, e2 := strconv.ParseFloat(field(rec, lonI), 64)
		if e1 != nil || e2 != nil {
			continue
		}
		f := Fire{Lat: lat, Lon: lon}
		if hasConf {
			f.Confidence = field(rec, confI)
		}
		if hasFRP {
			f.FRP, _ = strconv.ParseFloat(field(rec, frpI), 64)
		}
		// Filter marginal/low-intensity detections to keep the overlay legible
		// and the payload small.
		if f.Confidence == "l" {
			continue
		}
		if cfg.MinFRP > 0 && f.FRP < cfg.MinFRP {
			continue
		}
		out = append(out, f)
	}
	return out, nil
}

func field(rec []string, i int) string {
	if i >= 0 && i < len(rec) {
		return strings.TrimSpace(rec[i])
	}
	return ""
}
