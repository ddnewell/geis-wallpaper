package solar

import (
	"math"
	"testing"
	"time"
)

// At an equinox near 12:00 UTC the subsolar point is close to (0N, 0E):
// declination ~0 and solar noon near the prime meridian.
func TestSubsolarEquinox(t *testing.T) {
	when := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	p := SubsolarPoint(when)
	if math.Abs(p.SubLat) > 1.5 {
		t.Errorf("equinox declination = %.3f, want ~0", p.SubLat)
	}
	if math.Abs(p.SubLon) > 5 {
		t.Errorf("equinox subsolar lon = %.3f, want ~0", p.SubLon)
	}
}

// At the June solstice the declination is near +23.4 deg.
func TestSubsolarSolstice(t *testing.T) {
	when := time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC)
	p := SubsolarPoint(when)
	if math.Abs(p.SubLat-23.44) > 1.0 {
		t.Errorf("June solstice declination = %.3f, want ~23.44", p.SubLat)
	}
}

// At midnight UTC the subsolar longitude should be near the antimeridian.
func TestSubsolarMidnight(t *testing.T) {
	when := time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)
	p := SubsolarPoint(when)
	if math.Abs(math.Abs(p.SubLon)-180) > 5 {
		t.Errorf("midnight subsolar lon = %.3f, want ~±180", p.SubLon)
	}
}

// The sun should be high near the subsolar point and below the horizon on the
// opposite side of the globe.
func TestElevation(t *testing.T) {
	when := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)
	p := SubsolarPoint(when)
	atSub := p.Elevation(p.SubLon, p.SubLat)
	if atSub < 88 {
		t.Errorf("elevation at subsolar point = %.2f, want ~90", atSub)
	}
	anti := p.Elevation(wrapLon(p.SubLon+180), -p.SubLat)
	if anti > -80 {
		t.Errorf("elevation at antipode = %.2f, want ~-90", anti)
	}
}

func TestIrradiance(t *testing.T) {
	if g := GlobalHorizontalIrradiance(-5, 6); g != 0 {
		t.Errorf("GHI below horizon = %v, want 0", g)
	}
	noon := GlobalHorizontalIrradiance(90, 6)
	if noon < 900 || noon > 1300 {
		t.Errorf("GHI at zenith = %.1f, want ~1000-1200", noon)
	}
	low := GlobalHorizontalIrradiance(10, 6)
	if low >= noon {
		t.Errorf("low-sun GHI %.1f should be < zenith GHI %.1f", low, noon)
	}
}
