// Package feeds defines the normalized event data structures and the fetchers
// that produce them from upstream sources. The same fetchers are used by the
// client (direct, for keyless feeds) and the companion server (all feeds,
// including keyed ones whose keys stay server-side).
package feeds

import "time"

// Payload is the aggregated event data. It is what the companion serves as
// /payload.json and what the client caches locally; the renderer consumes it.
type Payload struct {
	GeneratedAt time.Time    `json:"generated_at"`
	Storms      []Storm      `json:"storms"`
	Quakes      []Quake      `json:"quakes"`
	Aurora      []AuroraCell `json:"aurora"`
	ISS         *ISS         `json:"iss,omitempty"`
	Ships       []Ship       `json:"ships"`
	Fires       []Fire       `json:"fires"`
}

// Storm is an active tropical cyclone with current state and forecast track.
// Category uses the Saffir-Simpson keying from the legacy icon set (-5..5):
//
//	-5 remnants, -4 invest, -3 extratropical, -2/-1 depression,
//	 0 tropical storm, 1..5 hurricane categories.
type Storm struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Category   int             `json:"category"`
	Lat        float64         `json:"lat"`
	Lon        float64         `json:"lon"`
	WindKt     float64         `json:"wind_kt"`
	PressureMb float64         `json:"pressure_mb"`
	Movement   string          `json:"movement"`
	Forecast   []ForecastPoint `json:"forecast"`
}

// ForecastPoint is a single predicted position along a storm's track.
type ForecastPoint struct {
	Hour     int     `json:"hour"`
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	Category int     `json:"category"`
}

// Quake is a single earthquake event.
type Quake struct {
	Mag     float64   `json:"mag"`
	Lat     float64   `json:"lat"`
	Lon     float64   `json:"lon"`
	DepthKm float64   `json:"depth_km"`
	Place   string    `json:"place"`
	Time    time.Time `json:"time"`
}

// AuroraCell is one cell of the OVATION aurora oval forecast.
type AuroraCell struct {
	Lon       float64 `json:"lon"`
	Lat       float64 `json:"lat"`
	Intensity int     `json:"intensity"` // 0..100 probability
}

// ISS is the current position of the International Space Station.
type ISS struct {
	Lat  float64   `json:"lat"`
	Lon  float64   `json:"lon"`
	Time time.Time `json:"time"`
}

// Ship is an AIS vessel position (keyed feed; populated server-side).
type Ship struct {
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
	Name string  `json:"name"`
	MMSI string  `json:"mmsi"`
}

// Fire is a NASA FIRMS active-fire detection (keyed feed; server-side).
type Fire struct {
	Lat        float64 `json:"lat"`
	Lon        float64 `json:"lon"`
	Confidence string  `json:"confidence"`
	FRP        float64 `json:"frp"` // fire radiative power (MW)
}
