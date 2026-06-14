package feeds

import (
	"context"
	"fmt"
	"time"
)

// SatelliteURL builds a WMS GetMap request for a full-globe equirectangular
// (EPSG:4326) cloud image at the master 2560x1280 grid. Both sources are
// keyless. NOTE the BBOX axis order differs by provider (verified working):
//   - NOAA nowCoast (GeoServer): BBOX = minLat,minLon,maxLat,maxLon
//   - NASA GIBS:                 BBOX = minLon,minLat,maxLon,maxLat
func SatelliteURL(layer string, t time.Time) string {
	switch layer {
	case "gibs_truecolor":
		// MODIS Aqua true color; use the previous UTC day for data availability.
		d := t.UTC().AddDate(0, 0, -1).Format("2006-01-02")
		return "https://gibs.earthdata.nasa.gov/wms/epsg4326/best/wms.cgi?" +
			"SERVICE=WMS&VERSION=1.3.0&REQUEST=GetMap" +
			"&LAYERS=MODIS_Aqua_CorrectedReflectance_TrueColor&FORMAT=image/png" +
			"&BBOX=-180,-90,180,90&WIDTH=2560&HEIGHT=1280&CRS=EPSG:4326&TIME=" + d
	default: // "nowcoast_global_longwave": global IR clouds, day & night.
		ts := t.UTC().Format("2006-01-02T15:04:05Z")
		return "https://nowcoast.noaa.gov/geoserver/satellite/wms?" +
			"SERVICE=WMS&VERSION=1.3.0&REQUEST=GetMap" +
			"&LAYERS=global_longwave_imagery_mosaic&FORMAT=image/png" +
			"&BBOX=-90,-180,90,180&WIDTH=2560&HEIGHT=1280&CRS=EPSG:4326&TIME=" + ts
	}
}

// FetchSatellite returns the cloud image PNG bytes for the given layer at time t.
func FetchSatellite(ctx context.Context, layer string, t time.Time) ([]byte, error) {
	b, err := GetBytes(ctx, SatelliteURL(layer, t))
	if err != nil {
		return nil, err
	}
	if len(b) < 1024 {
		return nil, fmt.Errorf("satellite image suspiciously small (%d bytes)", len(b))
	}
	return b, nil
}
