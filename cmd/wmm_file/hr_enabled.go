//go:build !wmm_no_hr

package main

import (
	"github.com/westphae/geomag/pkg/wmm"
	"github.com/westphae/geomag/pkg/wmm/wmmhr"
)

func hrModelLoader() *wmm.Model { return wmmhr.Default() }
