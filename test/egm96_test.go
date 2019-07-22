package main

import (
	"testing"

	"github.com/westphae/geomag/pkg/egm96"
	"github.com/westphae/geomag/pkg/units"
)

func TestEGM96GridLookup(t *testing.T) {
	lats := []units.Degrees{38, -12.25, -84.75, 26, 0}
	lngs := []units.Degrees{270, 82.75, 180.5, 279.5, 0}
	hts  := []units.Meters{-30.262, -67.347, -40.254, -26.621, 17.162}

	for i:=0; i<len(lats); i++ {
		_, p := egm96.GetEGM96GridPoint(units.Location{
			Latitude: lats[i],
			Longitude: lngs[i],
			Height: 0,
		})

		if dLat := p.Latitude - lats[i]; dLat < -EPS || dLat > EPS {
			t.Errorf("EGM96 Geoid height lookup changed the latitude from %6.4f to %6.4f", lats[i], p.Latitude)
		}
		if dLng := p.Longitude - lngs[i]; dLng < -EPS || dLng > EPS {
			t.Errorf("EGM96 Geoid height lookup changed the longitude from %6.4f to %6.4f", lngs[i], p.Longitude)
		}
		if dh := p.Height - hts[i]; dh < -EPS || dh > EPS {
			t.Errorf("EGM96 Geoid height incorrect, expected %6.4f, calculated %6.4f", hts[i], p.Height)
		}
	}
}

func TestEGM96GridInterpolation(t *testing.T) {
	lats := []units.Degrees{38, -12.25, 0, 38.628155, -14.621217, 46.874319, -23.617446, 38.625473, -0.466744}
	lngs := []units.Degrees{270, 82.75, 0, 269.779155, 305.021114, 102.448729, 133.874712, 359.999500, 0.002300}
	hts  := []units.Meters{-30.262, -67.347, 17.162, -31.628, -2.969, -43.575, 15.871, 50.066, 17.329}

	for i:=0; i<len(lats); i++ {
		_, h := egm96.CalculateHeightCorrection(units.Location{
			Latitude: lats[i],
			Longitude: lngs[i],
			Height: 0,
		})
		dh := h - hts[i]
		// 0.1 seems to be the error introduced by bi-linear interpolation rather than splines
		if dh < -0.1 || dh > 0.1 {
			t.Errorf("EGM96 Geoid height incorrect, expected %6.4f, calculated %6.4f", hts[i], h)
		}
	}
}


