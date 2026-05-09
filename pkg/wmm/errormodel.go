package wmm

// ErrorModel holds the global-average uncertainty values for a WMM model
// release, as published in NOAA's technical report for that release. The
// values change with each model (every five years), so they are bound to
// a particular *Model rather than to the package.
//
// X/Y/Z/H/F are global-average one-standard-deviation uncertainties in
// nanoteslas. I is in degrees. DA is the rough global-average declination
// uncertainty away from the magnetic poles, in degrees. DB is the H-field
// uncertainty scale used to compute the location-dependent declination
// uncertainty near the poles (see MagneticField.ErrD).
type ErrorModel struct {
	X, Y, Z, H, F float64
	I             float64
	DA            float64
	DB            float64
}

// defaultErrorModels maps COF names (the second whitespace-separated token
// on the WMM.COF header line — e.g. "WMM-2025", "WMM-2020") to their
// published error models. Add an entry here when bumping the embedded COF
// to a new release.
//
// For any COF whose name is not in this map, NewModel/LoadModel/ParseModel
// will leave the resulting *Model with a zero ErrorModel; callers can
// supply one explicitly via (*Model).SetErrorModel.
var defaultErrorModels = map[string]ErrorModel{
	"WMM-2025": {X: 137, Y: 89, Z: 141, H: 133, F: 138, I: 0.20, DA: 0.26, DB: 5417},
	"WMM-2020": {X: 131, Y: 94, Z: 157, H: 128, F: 148, I: 0.21, DA: 0.26, DB: 5625},
}
