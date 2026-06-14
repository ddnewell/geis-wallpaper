package solar

import "math"

// ASHRAE clear-sky model coefficients (21st of each month), converted to SI.
//
//	A = apparent extraterrestrial direct-normal flux (W/m^2)
//	B = atmospheric extinction coefficient (dimensionless)
//	C = diffuse-to-direct-normal ratio (dimensionless)
//
// Source: ASHRAE Handbook of Fundamentals clear-sky model.
var (
	ashraeA = [12]float64{1230, 1215, 1186, 1136, 1104, 1088, 1085, 1107, 1151, 1192, 1221, 1233}
	ashraeB = [12]float64{0.142, 0.144, 0.156, 0.180, 0.196, 0.205, 0.207, 0.201, 0.177, 0.160, 0.149, 0.142}
	ashraeC = [12]float64{0.058, 0.060, 0.071, 0.097, 0.121, 0.134, 0.136, 0.122, 0.092, 0.073, 0.063, 0.057}
)

// GlobalHorizontalIrradiance returns the modeled clear-sky global horizontal
// irradiance (W/m^2) for a given solar elevation (degrees) in the given month
// (1-12). Returns 0 at or below the horizon.
//
//	I_DN = A * exp(-B / sin(elev))      (direct normal)
//	GHI  = I_DN * sin(elev) + C * I_DN  (direct horizontal + diffuse)
func GlobalHorizontalIrradiance(elevationDeg float64, month int) float64 {
	if elevationDeg <= 0 {
		return 0
	}
	i := month - 1
	if i < 0 || i > 11 {
		i = 0
	}
	sinElev := math.Sin(degToRad(elevationDeg))
	if sinElev <= 1e-4 {
		return 0
	}
	directNormal := ashraeA[i] * math.Exp(-ashraeB[i]/sinElev)
	return directNormal*sinElev + ashraeC[i]*directNormal
}
