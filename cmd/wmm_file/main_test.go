package main

import (
	"bufio"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/westphae/geomag/pkg/wmm"
)

// TestAgainstNOAASample drives wmm_file's per-line processor with NOAA's
// official sample_coords.txt and compares the result column-by-column
// against NOAA's published sample_output.txt. K (km) and M (m) altitude
// rows must match within rounding (0.05 nT / 0.5 arc-min); F (feet) rows
// are tolerated to ≲0.5 nT to absorb NOAA's foot-to-km transcription error
// (3280.0839895 in wmm_file.c instead of the correct 3280.83989501…).
func TestAgainstNOAASample(t *testing.T) {
	in, err := os.Open("testdata/sample_coords.txt")
	if err != nil {
		t.Fatalf("open input: %v", err)
	}
	defer in.Close()

	expectedFile, err := os.Open("testdata/sample_output_noaa.txt")
	if err != nil {
		t.Fatalf("open expected: %v", err)
	}
	defer expectedFile.Close()
	expected := bufio.NewScanner(expectedFile)
	if !expected.Scan() {
		t.Fatalf("expected file empty")
	} // skip header

	model := wmm.Default()
	w := bufio.NewWriter(os.Stderr) // unused; processLine needs a writer
	_ = w

	scanner := bufio.NewScanner(in)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		raw := scanner.Text()
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		var actualBuf strings.Builder
		bw := bufio.NewWriter(&actualBuf)
		if err := processLine(bw, model, raw, line, false); err != nil {
			t.Errorf("line %d: processLine: %v", lineNo, err)
			continue
		}
		if err := bw.Flush(); err != nil {
			t.Errorf("line %d: flush: %v", lineNo, err)
			continue
		}
		actualLine := strings.TrimRight(actualBuf.String(), "\n")

		if !expected.Scan() {
			t.Fatalf("expected output truncated at line %d", lineNo)
		}
		expectedLine := expected.Text()

		// Echoed input columns 1-5 must match byte-for-byte (we echo the raw input tokens).
		actualEcho := strings.Join(strings.Fields(actualLine)[:5], " ")
		expectedEcho := strings.Join(strings.Fields(expectedLine)[:5], " ")
		if actualEcho != expectedEcho {
			t.Errorf("line %d: echoed input differs:\n  got:  %s\n  want: %s",
				lineNo, actualEcho, expectedEcho)
			continue
		}

		// Compare the 16 numeric output columns with tolerance.
		// Layout: D_deg D_min I_deg I_min H X Y Z F dD_min dI_min dH dX dY dZ dF
		// D_deg / D_min / I_deg / I_min are formatted as "16d 48m" — strip suffix.
		actualNums := stripUnits(strings.Fields(actualLine)[5:])
		expectedNums := stripUnits(strings.Fields(expectedLine)[5:])
		if len(actualNums) != len(expectedNums) {
			t.Errorf("line %d: column count differs: got %d, want %d",
				lineNo, len(actualNums), len(expectedNums))
			continue
		}
		// Distinguish F-altitude rows: tolerance is wider for those.
		altTok := strings.Fields(line)[2]
		feet := strings.HasPrefix(strings.ToUpper(altTok), "F")

		// Per-column tolerances.
		tolNT := 0.05
		if feet {
			tolNT = 0.5
		}
		tolMin := 0.5
		if feet {
			tolMin = 1.0
		}
		colTol := []float64{
			tolMin, tolMin, tolMin, tolMin, // D_deg, D_min, I_deg, I_min — degrees-and-minutes parts (these are integer or 0-decimal so tolerance is loose)
			tolNT, tolNT, tolNT, tolNT, tolNT, // H, X, Y, Z, F
			tolMin, tolMin, // dD_min, dI_min
			tolNT, tolNT, tolNT, tolNT, tolNT, // dH, dX, dY, dZ, dF
		}
		for k := range actualNums {
			a, errA := strconv.ParseFloat(actualNums[k], 64)
			e, errE := strconv.ParseFloat(expectedNums[k], 64)
			if errA != nil || errE != nil {
				t.Errorf("line %d col %d: parse: got %q, want %q",
					lineNo, k, actualNums[k], expectedNums[k])
				continue
			}
			if math.Abs(a-e) > colTol[k] {
				t.Errorf("line %d col %d: %.4f vs %.4f (tol %v)\n  got:  %s\n  want: %s",
					lineNo, k, a, e, colTol[k], actualLine, expectedLine)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan input: %v", err)
	}
}

// stripUnits removes the trailing "d" / "m" suffix NOAA uses on the four
// degrees-minutes columns, returning numeric strings.
func stripUnits(toks []string) []string {
	out := make([]string, len(toks))
	for i, t := range toks {
		out[i] = strings.TrimRight(t, "dm")
	}
	return out
}
