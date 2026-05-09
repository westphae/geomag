//go:build wmm_no_hr

package main

import (
	"fmt"
	"os"

	"github.com/westphae/geomag/pkg/wmm"
)

func hrModelLoader() *wmm.Model {
	fmt.Fprintln(os.Stderr,
		"--hr is unavailable in this build; rebuild without -tags wmm_no_hr "+
			"(or run `make install`) to enable HR support.")
	os.Exit(2)
	return nil
}
