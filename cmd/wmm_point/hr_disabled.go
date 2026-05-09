//go:build wmm_no_hr

package main

import (
	"fmt"
	"os"

	"github.com/westphae/geomag/pkg/wmm"
)

// hrModelLoader is the build-tag-disabled stub that runs when this binary
// was compiled with `-tags wmm_no_hr` (the lean build that omits the
// embedded WMMHR coefficients to save ~530 KB). Calling --hr on a lean
// binary prints a clear error and exits non-zero.
func hrModelLoader() *wmm.Model {
	fmt.Fprintln(os.Stderr,
		"--hr is unavailable in this build; rebuild without -tags wmm_no_hr "+
			"(or run `make install`) to enable HR support.")
	os.Exit(2)
	return nil
}
