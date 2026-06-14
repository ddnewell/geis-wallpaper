package feeds

import (
	"context"
)

// ovationURL is the keyless NOAA SWPC OVATION aurora-oval nowcast.
const ovationURL = "https://services.swpc.noaa.gov/json/ovation_aurora_latest.json"

// auroraMinIntensity filters out the vast majority of zero/near-zero cells so
// the payload stays small; only cells with meaningful aurora probability are kept.
const auroraMinIntensity = 2

type ovationResponse struct {
	// Coordinates is a list of [longitude(0..359), latitude(-90..90), aurora(0..100)].
	Coordinates [][3]float64 `json:"coordinates"`
}

// FetchAurora returns the aurora oval cells with intensity >= auroraMinIntensity.
// Longitudes are normalized from 0..360 to -180..180.
func FetchAurora(ctx context.Context) ([]AuroraCell, error) {
	var r ovationResponse
	if err := GetJSON(ctx, ovationURL, &r); err != nil {
		return nil, err
	}
	out := make([]AuroraCell, 0, 4096)
	for _, c := range r.Coordinates {
		intensity := int(c[2])
		if intensity < auroraMinIntensity {
			continue
		}
		lon := c[0]
		if lon > 180 {
			lon -= 360
		}
		out = append(out, AuroraCell{Lon: lon, Lat: c[1], Intensity: intensity})
	}
	return out, nil
}
