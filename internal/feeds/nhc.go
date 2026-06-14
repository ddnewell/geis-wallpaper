package feeds

import (
	"context"
	"strconv"
	"strings"
)

// currentStormsURL is the keyless NHC feed of active tropical cyclones (all
// NHC basins). Returns {"activeStorms": []} when none are active.
const currentStormsURL = "https://www.nhc.noaa.gov/CurrentStorms.json"

type nhcResponse struct {
	ActiveStorms []nhcStorm `json:"activeStorms"`
}

type nhcStorm struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Classification   string   `json:"classification"`
	Intensity        string   `json:"intensity"` // sustained wind, kt (string)
	Pressure         string   `json:"pressure"`  // mb (string)
	Latitude         string   `json:"latitude"`  // e.g. "16.3N"
	Longitude        string   `json:"longitude"` // e.g. "60.5W"
	LatitudeNumeric  *float64 `json:"latitudeNumeric"`
	LongitudeNumeric *float64 `json:"longitudeNumeric"`
	MovementDir      any      `json:"movementDir"`
	MovementSpeed    any      `json:"movementSpeed"`
}

// FetchStorms returns active tropical cyclones from the NHC. Forecast tracks
// are not included inline by this feed; markers use current positions.
func FetchStorms(ctx context.Context) ([]Storm, error) {
	var r nhcResponse
	if err := GetJSON(ctx, currentStormsURL, &r); err != nil {
		return nil, err
	}
	out := make([]Storm, 0, len(r.ActiveStorms))
	for _, s := range r.ActiveStorms {
		lat, lon := stormLatLon(s)
		wind, _ := strconv.ParseFloat(strings.TrimSpace(s.Intensity), 64)
		pres, _ := strconv.ParseFloat(strings.TrimSpace(s.Pressure), 64)
		out = append(out, Storm{
			ID:         s.ID,
			Name:       strings.TrimSpace(s.Name),
			Category:   saffirSimpson(s.Classification, wind),
			Lat:        lat,
			Lon:        lon,
			WindKt:     wind,
			PressureMb: pres,
		})
	}
	return out, nil
}

func stormLatLon(s nhcStorm) (lat, lon float64) {
	if s.LatitudeNumeric != nil {
		lat = *s.LatitudeNumeric
	} else {
		lat = parseCoord(s.Latitude, 'N', 'S')
	}
	if s.LongitudeNumeric != nil {
		lon = *s.LongitudeNumeric
	} else {
		lon = parseCoord(s.Longitude, 'E', 'W')
	}
	return lat, lon
}

// parseCoord parses "16.3N" / "60.5W" into a signed decimal degree.
func parseCoord(s string, pos, neg byte) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	sign := 1.0
	last := s[len(s)-1]
	if last == neg || last == neg+32 {
		sign = -1
	}
	if last == pos || last == neg || last == pos+32 || last == neg+32 {
		s = s[:len(s)-1]
	}
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return sign * v
}

// saffirSimpson maps an NHC classification + sustained wind (kt) to the legacy
// icon category (-5..5).
func saffirSimpson(classification string, windKt float64) int {
	switch strings.ToUpper(strings.TrimSpace(classification)) {
	case "HU", "MH", "TY", "STY": // hurricane / typhoon
		switch {
		case windKt >= 137:
			return 5
		case windKt >= 113:
			return 4
		case windKt >= 96:
			return 3
		case windKt >= 83:
			return 2
		default:
			return 1
		}
	case "TS", "STS": // (sub)tropical storm
		return 0
	case "TD", "SD": // (sub)tropical depression
		return -1
	case "PTC": // potential tropical cyclone
		return -2
	case "EX", "PT": // extratropical / post-tropical
		return -3
	case "INVEST", "DB": // invest / disturbance
		return -4
	case "LO", "REMNANTS", "RL": // remnant low
		return -5
	default:
		if windKt >= 64 {
			return 1
		}
		return 0
	}
}
