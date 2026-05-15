// wmm_file batch-processes a coordinate file into magnetic field values.
//
// Usage:
//
//	wmm_file [--cof_file=PATH] f [e|Errors|--Errors] INPUT_FILE OUTPUT_FILE
//
// The arg layout matches NOAA's reference C distribution. The leading "f"
// is required (NOAA's distribution reserves it for "file mode"; we accept
// it for compatibility but do nothing with it). Adding "e", "Errors", or
// "--Errors" between the "f" and INPUT_FILE appends seven uncertainty
// columns to each output row.
//
// Input format, one record per line:
//
//	<date> <coord-system> <altitude> <lat> <lng>
//
// where:
//   - <date> is a decimal year, e.g. 2025.5
//   - <coord-system> is "E" (height above WGS84 ellipsoid) or "M" (height
//     above mean sea level)
//   - <altitude> is a single-character unit prefix (K = kilometers,
//     M = meters, F = feet) followed by a signed decimal magnitude.
//     Example: "K100", "M1042", "F30000", "K-1", "K1.3"
//   - <lat> and <lng> are decimal degrees (lat positive north, lng positive
//     east; longitudes in either [-180,180] or [0,360] are accepted).
//
// The output is space-separated columns matching NOAA's reference output.
//
// Example NOAA-distributed sample input (sample_coords.txt):
//
//	2025.5 E K100  70.3 30.8
//	2026.5 E M1042 70.3 30.8
//	2027.5 E F30000 70.3 30.8
//
// Output is byte-for-byte identical to NOAA's reference wmm_file for K and
// M altitude inputs. For F (feet) inputs, our output may differ from
// NOAA's by ≲0.1 nT in some components: NOAA's reference C source uses a
// feet-to-km divisor of 3280.0839895 (a transcription error in
// wmm_file.c:606 — the correct value is 1000/0.3048 = 3280.83989501…),
// producing a ~2 m altitude error at 30000 ft. We use the exact
// international-foot conversion 0.3048. The discrepancy is far below the
// WMM error model's stated uncertainty (138 nT for F at WMM2025) and
// matters only for byte-comparison of CLI output, not for accuracy.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/westphae/geomag/internal/parsing"
	"github.com/westphae/geomag/pkg/egm96"
	"github.com/westphae/geomag/pkg/wmm"
)

const usage = "wmm_file [--hr | --cof_file=PATH] f [e|Errors|--Errors] INPUT_FILE OUTPUT_FILE"

const usageHeader = `wmm_file — batch-process a coordinate file into magnetic-field values.

Reads one geodetic record per line from INPUT_FILE and writes a space-
separated row of computed field values (D, I, H, X, Y, Z, F and their
secular variations) per record to OUTPUT_FILE. Argument layout matches
NOAA's reference C distribution (` + "`" + `wmm_file f INPUT OUTPUT` + "`" + `).

Usage:
  wmm_file [flags] f [e|Errors|--Errors] INPUT_FILE OUTPUT_FILE

Positional arguments:
  f                    Required literal — NOAA-compat marker for "file mode".
  e|Errors|--Errors    Optional. When present, appends seven uncertainty
                       columns (errD, errI, errH, errX, errY, errZ, errF).
  INPUT_FILE           Path to coordinate records, one per line.
  OUTPUT_FILE          Destination for computed-field rows. Overwritten if
                       it exists. A header row is written first.

Input record format (whitespace-separated):
  <date> <coord-system> <altitude> <lat> <lng>

  <date>          Decimal year (e.g. 2025.5).
  <coord-system>  "E" = height above WGS-84 ellipsoid, "M" = height above
                  mean sea level.
  <altitude>      Single-character unit prefix + signed decimal magnitude:
                  K = kilometers, M = meters, F = feet. e.g. K100, M1042,
                  F30000, K-1, K1.3.
  <lat>, <lng>    Decimal degrees. lat positive north, lng positive east
                  (longitudes in [-180,180] or [0,360] are both accepted).

Blank lines and lines starting with "#" are skipped.

Flags:`

const usageExamples = `
Examples:
  wmm_file f sample_coords.txt out.txt
  wmm_file f e sample_coords.txt out.txt          # include uncertainty cols
  wmm_file --hr f sample_coords.txt out.txt       # WMMHR2025 (degree 133)
  wmm_file --cof_file=WMM2020.COF f in.txt out.txt
`

var (
	headerNoErrors = "Date Coord-System Altitude Latitude Longitude" +
		" D_deg D_min I_deg I_min H_nT X_nT Y_nT Z_nT F_nT" +
		" dD_min dI_min dH_nT dX_nT dY_nT dZ_nT dF_nT"
	headerErrorsSuffix = " errD_min errI_min errH_nT errX_nT errY_nT errZ_nT errF_nT"
)

func main() {
	cofFile := flag.String("cof_file", "", "alternate WMM coefficients file (optional; default uses embedded WMM2025)")
	useHR := flag.Bool("hr", false, "use the high-resolution WMMHR model (degree 133); mutually exclusive with --cof_file")
	flag.Usage = func() {
		_, _ = fmt.Fprintln(os.Stderr, usageHeader)
		flag.PrintDefaults()
		_, _ = fmt.Fprint(os.Stderr, usageExamples)
	}
	flag.Parse()
	if *useHR && *cofFile != "" {
		_, _ = fmt.Fprintln(os.Stderr, "--hr and --cof_file are mutually exclusive")
		os.Exit(2)
	}

	args := flag.Args()
	withErrors := false
	switch len(args) {
	case 3:
		// f IN OUT
	case 4:
		// f (e|Errors|--Errors) IN OUT
		switch args[1] {
		case "e", "Errors", "--Errors":
			withErrors = true
		default:
			_, _ = fmt.Fprintf(os.Stderr, "unknown errors flag %q (expected e, Errors, or --Errors)\n%s\n", args[1], usage)
			os.Exit(2)
		}
		// Drop the errors token so positionals collapse to [f, IN, OUT].
		args = append(args[:1], args[2:]...)
	default:
		_, _ = fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
	if args[0] != "f" {
		_, _ = fmt.Fprintf(os.Stderr, "first argument must be \"f\" (got %q)\n%s\n", args[0], usage)
		os.Exit(2)
	}
	inputPath, outputPath := args[1], args[2]

	var model *wmm.Model
	switch {
	case *useHR:
		model = hrModelLoader()
	case *cofFile != "":
		var err error
		if model, err = wmm.LoadModel(*cofFile); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error loading WMM coefficients: %v\n", err)
			os.Exit(1)
		}
	default:
		model = wmm.Default()
	}

	in, err := os.Open(inputPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error opening input %s: %v\n", inputPath, err)
		os.Exit(1)
	}
	defer func() { _ = in.Close() }()

	out, err := os.Create(outputPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error creating output %s: %v\n", outputPath, err)
		os.Exit(1)
	}
	defer func() { _ = out.Close() }()

	w := bufio.NewWriter(out)
	defer func() { _ = w.Flush() }()

	header := headerNoErrors
	if withErrors {
		header += headerErrorsSuffix
	}
	if _, err := fmt.Fprintln(w, header); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error writing header: %v\n", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(in)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		raw := scanner.Text()
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if err := processLine(w, model, raw, line, withErrors); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "line %d: %v\n", lineNo, err)
			os.Exit(1)
		}
	}
	if err := scanner.Err(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error reading %s: %v\n", inputPath, err)
		os.Exit(1)
	}
}

func processLine(w *bufio.Writer, model *wmm.Model, raw, line string, withErrors bool) error {
	tokens := strings.Fields(line)
	if len(tokens) != 5 {
		return fmt.Errorf("expected 5 fields (date coord-system altitude lat lng), got %d in %q", len(tokens), raw)
	}
	dateTok, coordSys, altTok, latTok, lngTok := tokens[0], tokens[1], tokens[2], tokens[3], tokens[4]

	date, err := strconv.ParseFloat(dateTok, 64)
	if err != nil {
		return fmt.Errorf("invalid date %q: %w", dateTok, err)
	}
	height, err := parsing.ParseNOAAAltitude(altTok)
	if err != nil {
		return fmt.Errorf("invalid altitude: %w", err)
	}
	lat, err := strconv.ParseFloat(latTok, 64)
	if err != nil {
		return fmt.Errorf("invalid latitude %q: %w", latTok, err)
	}
	lng, err := strconv.ParseFloat(lngTok, 64)
	if err != nil {
		return fmt.Errorf("invalid longitude %q: %w", lngTok, err)
	}

	var loc egm96.Location
	switch coordSys {
	case "E":
		loc = egm96.NewLocationGeodetic(lat, lng, height)
	case "M":
		if loc, err = egm96.NewLocationMSL(lat, lng, height); err != nil {
			return fmt.Errorf("MSL location: %w", err)
		}
	default:
		return fmt.Errorf("coord-system %q not in {E, M}", coordSys)
	}

	mf, _ := model.MagneticField(loc, wmm.DecimalYear(date).ToTime())
	// model.MagneticField may return an "outside validity period" error along
	// with a usable result. wmm_file echoes whatever the math produces; the
	// validity warning would be noise per-line.

	x, y, z, dx, dy, dz := mf.Ellipsoidal()

	d := mf.D()
	i := mf.I()
	ddeg, dmin := splitDM(d)
	ideg, imin := splitDM(i)

	// Echo the input line verbatim, then the computed columns. NOAA's
	// reference uses a specific column-padded format string (note the
	// trailing space after the last %s); we match it for byte-for-byte
	// compatibility with NOAA's published sample_output.txt.
	if _, err := fmt.Fprintf(w, "%s %s %s %s %s ",
		dateTok, coordSys, altTok, latTok, lngTok); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, " %4dd %2.0fm  %4dd %2.0fm  %8.1f %8.1f %8.1f %8.1f %8.1f",
		ddeg, dmin, ideg, imin, mf.H(), x, y, z, mf.F()); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, " %7.1f   %7.1f     %8.1f %8.1f %8.1f %8.1f %8.1f",
		mf.DD()*60, mf.DI()*60, mf.DH(), dx, dy, dz, mf.DF()); err != nil {
		return err
	}
	if withErrors {
		if _, err := fmt.Fprintf(w, " %3.0f  %3.0f  %8.1f %8.1f %8.1f %8.1f %8.1f",
			mf.ErrD()*60, mf.ErrI()*60, mf.ErrH(), mf.ErrX(), mf.ErrY(), mf.ErrZ(), mf.ErrF()); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	return nil
}

// splitDM splits a decimal-degree value into the integer-degrees and
// fractional-arcminutes parts NOAA's wmm_file output uses. The minutes
// component is always non-negative; the sign lives on the degree component.
// Edge case: for |x| < 1 with x < 0, the integer part is 0 (trunc-toward-
// zero) and the sign is lost — that's NOAA's convention too.
func splitDM(x float64) (deg int, min float64) {
	deg = int(x)
	min = math.Abs(x-float64(deg)) * 60
	return deg, min
}
