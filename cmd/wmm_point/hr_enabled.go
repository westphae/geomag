//go:build !wmm_no_hr

package main

import (
	"github.com/westphae/geomag/pkg/wmm"
	"github.com/westphae/geomag/pkg/wmm/wmmhr"
)

// hrModelLoader returns the embedded WMMHR2025 model. Importing the wmmhr
// sub-package is what pulls the ~530 KB WMMHR.COF into the binary; the
// `//go:build !wmm_no_hr` tag at the top of this file lets users build
// without that import via `go build -tags wmm_no_hr ./...` (or `make
// install-lean`), in which case the HR loader in hr_disabled.go is
// compiled instead.
func hrModelLoader() *wmm.Model { return wmmhr.Default() }
