// Package solar computes the present-time solar geometry and clear-sky
// irradiance used to shade the map. Everything here is pure math (no network,
// no dependencies), which is what makes the wallpaper update work offline.
//
// The position model is the standard NOAA approximation (general solar
// position algorithm): fractional-year -> equation of time + declination ->
// subsolar point. Accuracy is ~0.01 deg, far finer than the ~0.07 deg/pixel of
// a 2560-wide map. All inputs are interpreted in UTC.
package solar

import (
	"math"
	"time"
)

// Position holds the subsolar point (where the sun is at the zenith) for an
// instant, in degrees.
type Position struct {
	SubLat float64 // = solar declination
	SubLon float64 // longitude of solar noon, in [-180,180]
}

// SubsolarPoint computes the subsolar point for time t (evaluated in UTC).
func SubsolarPoint(t time.Time) Position {
	ut := t.UTC()

	// Day-of-year (1-based) and fractional UTC hour.
	doy := ut.YearDay()
	hour := float64(ut.Hour()) + float64(ut.Minute())/60.0 + float64(ut.Second())/3600.0

	// Fractional year (radians), per the NOAA algorithm.
	daysInYear := 365.0
	if isLeap(ut.Year()) {
		daysInYear = 366.0
	}
	gamma := 2.0 * math.Pi / daysInYear * (float64(doy-1) + (hour-12.0)/24.0)

	// Equation of time, in minutes.
	eqTime := 229.18 * (0.000075 +
		0.001868*math.Cos(gamma) -
		0.032077*math.Sin(gamma) -
		0.014615*math.Cos(2*gamma) -
		0.040849*math.Sin(2*gamma))

	// Solar declination, in radians.
	decl := 0.006918 -
		0.399912*math.Cos(gamma) +
		0.070257*math.Sin(gamma) -
		0.006758*math.Cos(2*gamma) +
		0.000907*math.Sin(2*gamma) -
		0.002697*math.Cos(3*gamma) +
		0.001480*math.Sin(3*gamma)

	// Subsolar longitude (east-positive): the meridian where it is solar noon.
	// Derivation: true solar time (min) = UTC_min + eqTime + 4*lon; noon when = 720.
	//   => lon = (720 - UTC_min - eqTime) / 4 = 180 - 15*hour - eqTime/4
	subLon := 180.0 - 15.0*hour - eqTime/4.0
	subLon = wrapLon(subLon)

	return Position{
		SubLat: radToDeg(decl),
		SubLon: subLon,
	}
}

// Elevation returns the solar elevation angle (degrees above the horizon) at
// the point (lon,lat) for the given subsolar position. Negative below horizon.
func (p Position) Elevation(lon, lat float64) float64 {
	latR := degToRad(lat)
	subLatR := degToRad(p.SubLat)
	hourAngle := degToRad(lon - p.SubLon)
	sinElev := math.Sin(latR)*math.Sin(subLatR) +
		math.Cos(latR)*math.Cos(subLatR)*math.Cos(hourAngle)
	sinElev = clamp(sinElev, -1, 1)
	return radToDeg(math.Asin(sinElev))
}

func isLeap(y int) bool { return y%4 == 0 && (y%100 != 0 || y%400 == 0) }

func wrapLon(lon float64) float64 {
	for lon > 180 {
		lon -= 360
	}
	for lon < -180 {
		lon += 360
	}
	return lon
}

func degToRad(d float64) float64 { return d * math.Pi / 180.0 }
func radToDeg(r float64) float64 { return r * 180.0 / math.Pi }

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
