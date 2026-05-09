// wmm_grid generates a magnetic-field value over a 4-D grid in latitude,
// longitude, altitude, and time, evaluating the chosen output element at
// each grid point and writing one row per point to stdout or a file.
//
// Modeled after the wmm_grid command in NOAA's reference C distribution.
// Where NOAA's tool is interactive, this one is flag-driven so it composes
// naturally with shell pipelines and scripts; collapsing an axis (e.g. for a
// profile or time series) is done by setting that axis's min and max equal,
// matching NOAA's convention.
//
// Usage:
//
//	wmm_grid \
//	    --lat=START,END,STEP --lng=START,END,STEP \
//	    --alt=START,END,STEP --date=START,END,STEP \
//	    [--element=D] [--errors] [--output=FILE] [--cof_file=PATH]
//
// Each axis flag takes a comma-separated triple of (start, end, step). All
// values are in their natural units: latitudes/longitudes in decimal
// degrees, altitudes in kilometers above the WGS84 ellipsoid, dates in
// decimal years. Step must divide (end-start) evenly modulo float
// rounding; we step from start through end inclusive.
//
// To collapse an axis, set start=end (step is ignored). For example,
// --alt=0,0,0 holds altitude at sea level, --date=2026,2026,0 picks a
// single instant.
//
// --element selects which scalar value to emit. Supported names mirror the
// MagneticField getters: D (declination, deg), I (inclination, deg),
// F (total intensity, nT), H (horizontal intensity, nT), X/Y/Z (north/
// east/down components, nT), GV (grid variation, deg, only meaningful at
// |lat| > 55°), and the secular-variation versions dD (deg/yr), dI
// (deg/yr), dF (nT/yr), dH (nT/yr), dX/dY/dZ (nT/yr).
//
// --errors appends one extra column with the WMM uncertainty for that
// element (e.g. ErrF for --element=F).
//
// Output is one whitespace-separated row per grid point:
//
//	<lat> <lng> <alt> <date> <value> [<error>]
package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/westphae/geomag/pkg/egm96"
	"github.com/westphae/geomag/pkg/wmm"
)

const usage = `wmm_grid --lat=START,END,STEP --lng=START,END,STEP --alt=START,END,STEP --date=START,END,STEP [--element=D] [--errors] [--output=FILE] [--cof_file=PATH]`

// elementFn extracts a scalar value (or scalar value + uncertainty) from a
// computed MagneticField. value is what's emitted as the main column; err
// is the WMM uncertainty for that element when --errors is set.
type elementFn struct {
	name  string
	value func(wmm.MagneticField) float64
	err   func(wmm.MagneticField) float64 // may be nil if no published uncertainty
}

var elements = map[string]elementFn{
	"D":  {"D", func(m wmm.MagneticField) float64 { return m.D() }, func(m wmm.MagneticField) float64 { return m.ErrD() }},
	"I":  {"I", func(m wmm.MagneticField) float64 { return m.I() }, func(m wmm.MagneticField) float64 { return m.ErrI() }},
	"F":  {"F", func(m wmm.MagneticField) float64 { return m.F() }, func(m wmm.MagneticField) float64 { return m.ErrF() }},
	"H":  {"H", func(m wmm.MagneticField) float64 { return m.H() }, func(m wmm.MagneticField) float64 { return m.ErrH() }},
	"X":  {"X", func(m wmm.MagneticField) float64 { x, _, _, _, _, _ := m.Ellipsoidal(); return x }, func(m wmm.MagneticField) float64 { return m.ErrX() }},
	"Y":  {"Y", func(m wmm.MagneticField) float64 { _, y, _, _, _, _ := m.Ellipsoidal(); return y }, func(m wmm.MagneticField) float64 { return m.ErrY() }},
	"Z":  {"Z", func(m wmm.MagneticField) float64 { _, _, z, _, _, _ := m.Ellipsoidal(); return z }, func(m wmm.MagneticField) float64 { return m.ErrZ() }},
	"GV": {"GV", nil, nil}, // populated below; needs the Location too
	"dD": {"dD", func(m wmm.MagneticField) float64 { return m.DD() }, nil},
	"dI": {"dI", func(m wmm.MagneticField) float64 { return m.DI() }, nil},
	"dF": {"dF", func(m wmm.MagneticField) float64 { return m.DF() }, nil},
	"dH": {"dH", func(m wmm.MagneticField) float64 { return m.DH() }, nil},
	"dX": {"dX", func(m wmm.MagneticField) float64 { _, _, _, dx, _, _ := m.Ellipsoidal(); return dx }, nil},
	"dY": {"dY", func(m wmm.MagneticField) float64 { _, _, _, _, dy, _ := m.Ellipsoidal(); return dy }, nil},
	"dZ": {"dZ", func(m wmm.MagneticField) float64 { _, _, _, _, _, dz := m.Ellipsoidal(); return dz }, nil},
}

// axis holds a parsed (start, end, step) triple for one of the four sweep
// dimensions. start may equal end to collapse the axis.
type axis struct {
	start, end, step float64
	name             string // for error messages
}

// parseAxis parses "start,end,step" into an axis. step is ignored when
// start == end.
func parseAxis(name, raw string) (axis, error) {
	a := axis{name: name}
	parts := strings.Split(raw, ",")
	if len(parts) != 3 {
		return a, fmt.Errorf("--%s expects START,END,STEP (got %q)", name, raw)
	}
	var err error
	if a.start, err = strconv.ParseFloat(parts[0], 64); err != nil {
		return a, fmt.Errorf("--%s start: %w", name, err)
	}
	if a.end, err = strconv.ParseFloat(parts[1], 64); err != nil {
		return a, fmt.Errorf("--%s end: %w", name, err)
	}
	if a.step, err = strconv.ParseFloat(parts[2], 64); err != nil {
		return a, fmt.Errorf("--%s step: %w", name, err)
	}
	if a.end < a.start {
		return a, fmt.Errorf("--%s: end (%v) is less than start (%v)", name, a.end, a.start)
	}
	if a.start != a.end && a.step <= 0 {
		return a, fmt.Errorf("--%s: step must be positive when start != end (got %v)", name, a.step)
	}
	return a, nil
}

// values yields each sample point along the axis, inclusive of both ends.
// A small epsilon absorbs float-step rounding.
func (a axis) values() []float64 {
	if a.start == a.end {
		return []float64{a.start}
	}
	const epsRel = 1e-9
	n := int(math.Floor((a.end-a.start)/a.step+epsRel)) + 1
	out := make([]float64, n)
	for i := range n {
		out[i] = a.start + float64(i)*a.step
	}
	// Snap the final sample exactly to end to absorb step accumulation.
	out[n-1] = a.end
	return out
}

func main() {
	latRaw := flag.String("lat", "", "latitude axis: START,END,STEP (decimal degrees)")
	lngRaw := flag.String("lng", "", "longitude axis: START,END,STEP (decimal degrees)")
	altRaw := flag.String("alt", "", "altitude axis: START,END,STEP (kilometers above WGS84 ellipsoid)")
	dateRaw := flag.String("date", "", "time axis: START,END,STEP (decimal years)")
	element := flag.String("element", "D", "output element: D, I, F, H, X, Y, Z, GV, dD, dI, dF, dH, dX, dY, dZ")
	withErrors := flag.Bool("errors", false, "append the WMM uncertainty for the chosen element as a second column")
	outPath := flag.String("output", "", "write rows to this file (default: stdout)")
	cofFile := flag.String("cof_file", "", "alternate WMM coefficients file (default: embedded WMM2025)")
	flag.Usage = func() {
		_, _ = fmt.Fprintln(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.Parse()

	if *latRaw == "" || *lngRaw == "" || *altRaw == "" || *dateRaw == "" {
		_, _ = fmt.Fprintln(os.Stderr, "all four axes (--lat, --lng, --alt, --date) are required\n"+usage)
		os.Exit(2)
	}

	latAxis, err := parseAxis("lat", *latRaw)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	lngAxis, err := parseAxis("lng", *lngRaw)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	altAxis, err := parseAxis("alt", *altRaw)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	dateAxis, err := parseAxis("date", *dateRaw)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if *element == "GV" {
		_, _ = fmt.Fprintln(os.Stderr, "--element=GV is not yet supported by wmm_grid (it varies with location, not just the field). File an issue if you need it.")
		os.Exit(2)
	}
	el, ok := elements[*element]
	if !ok {
		_, _ = fmt.Fprintf(os.Stderr, "unknown --element %q (try one of: D, I, F, H, X, Y, Z, dD, dI, dF, dH, dX, dY, dZ)\n", *element)
		os.Exit(2)
	}
	if *withErrors && el.err == nil {
		_, _ = fmt.Fprintf(os.Stderr, "--errors is not available for --element=%s (no published uncertainty for secular-variation components)\n", *element)
		os.Exit(2)
	}

	model, err := loadModel(*cofFile)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "loading WMM coefficients: %v\n", err)
		os.Exit(1)
	}

	out := os.Stdout
	if *outPath != "" {
		out, err = os.Create(*outPath)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "creating output: %v\n", err)
			os.Exit(1)
		}
		defer func() { _ = out.Close() }()
	}
	w := bufio.NewWriter(out)
	defer func() { _ = w.Flush() }()

	if err := emitGrid(w, model, latAxis, lngAxis, altAxis, dateAxis, el, *withErrors); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "emit: %v\n", err)
		os.Exit(1)
	}
}

func loadModel(path string) (*wmm.Model, error) {
	if path == "" {
		return wmm.Default(), nil
	}
	return wmm.LoadModel(path)
}

func emitGrid(w *bufio.Writer, model *wmm.Model, latAxis, lngAxis, altAxis, dateAxis axis, el elementFn, withErrors bool) error {
	// Iteration order: time innermost (cheapest, hits the per-location cache),
	// then altitude, latitude, longitude — matching the cache-friendly order
	// documented in pkg/wmm.
	for _, lng := range lngAxis.values() {
		for _, lat := range latAxis.values() {
			for _, alt := range altAxis.values() {
				loc := egm96.NewLocationGeodetic(lat, lng, alt*1000) // km → m
				for _, date := range dateAxis.values() {
					t := wmm.DecimalYear(date).ToTime()
					mf, _ := model.MagneticField(loc, t)
					val := el.value(mf)
					if withErrors {
						if _, err := fmt.Fprintf(w, "%g %g %g %g %g %g\n",
							lat, lng, alt, date, val, el.err(mf)); err != nil {
							return err
						}
					} else {
						if _, err := fmt.Fprintf(w, "%g %g %g %g %g\n",
							lat, lng, alt, date, val); err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return nil
}
