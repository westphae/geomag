package main

import (
	"testing"

	"github.com/westphae/geomag/pkg/polynomial"
)

const EPS = 1e-6

func TestPow(t *testing.T) {
	var (
		xs = []float64{2.0, 0.5, 1.0, 3.14, 10}
		ns = []int{5, 3, 4, 0, -3}
		ys = []float64{32, 0.125, 1, 1, 0.001}
	)

	for i:=0; i<len(xs); i++ {
		y := polynomial.Pow(xs[i], ns[i])
		dy := y - ys[i]
		if dy < -EPS || dy > EPS {
			t.Errorf("Pow expected %4.1f, calculated %4.1f", ys[i], y)
		}
	}
}

func TestFactorial(t *testing.T) {
	var (
		ns = []int{5, 3, 4, 0, 1}
		zs = []int{120, 6, 24, 1, 1}
	)

	for i:=0; i<len(ns); i++ {
		z := polynomial.Factorial(ns[i])
		dz := z - zs[i]
		if dz != 0 {
			t.Errorf("Factorial expected %d, calculated %d", zs[i], z)
		}
	}
}

func TestEvaluate(t *testing.T) {
	var (
		cs = [][]float64{
			{-1, 0, 2},
			{0.5, -1, 1, 2},
		}
		xs = []float64{2, 0.5, -1.5}
		ys = [][]float64{
			{7, 18.5},
			{-0.5, 0.5},
			{3.5, -2.5},
		}
	)

	for i:=0; i<len(cs); i++ {
		for j:=0; j<len(xs); j++ {
			p := polynomial.NewPolynomial(cs[i])
			y := p.Evaluate(xs[j])
			dy := y - ys[j][i]
			if dy < -EPS || dy > EPS {
				t.Errorf("Evaluate for %v(%3.1f) expected %4.1f, calculated %4.1f",
					cs[i], xs[j], ys[j][i], y)
			}
		}
	}
}

func TestDerivative(t *testing.T) {
	var (
		cs = [][]float64{
			{-1, 0, 2},
			{0.5, -1, 1, 2},
		}
		ds = [][]float64{
			{0, 4},
			{-1, 2, 6},
		}
		dds = [][]float64{
			{4},
			{2, 12},
		}
	)

	for i:=0; i<len(cs); i++ {
		p := polynomial.NewPolynomial(cs[i])

		y := p.Derivative(1).Coefficients()
		for j, d := range ds[i] {
			dy := y[j]-d
			if dy < -EPS || dy > EPS {
				t.Errorf("Derivative of %v was wrong, expecting %v, got %v",
					cs[i], ds[i], y)
			}
		}

		y = p.Derivative(2).Coefficients()
		for j, d := range dds[i] {
			dy := y[j]-d
			if dy < -EPS || dy > EPS {
				t.Errorf("Second derivative of %v was wrong, expecting %v, got %v",
					cs[i], dds[i], y)
			}
		}
	}
}

func TestLegendrePolynomials(t *testing.T) {
	cs := [][]float64{
		{1},
		{0, 1},
		{-1.0/2, 0, 3.0/2},
		{0, -3.0/2, 0, 5.0/2},
		{3.0/8, 0, -30.0/8, 0, 35.0/8},
		{0, 15.0/8, 0, -70.0/8, 0, 63.0/8},
		{-5.0/16, 0, 105.0/16, 0, -315.0/16, 0, 231.0/16},
		{0, -35.0/16, 0, 315.0/16, 0, -693.0/16, 0, 429.0/16},
		{35.0/128, 0, -1260.0/128, 0, 6930.0/128, 0, -12012.0/128, 0, 6435.0/128},
	}

	for n, cExpected := range cs {
		cCalculated := polynomial.LegendrePolynomial(n).Coefficients()
		for j:=0; j<=n; j++ {
			dc := cCalculated[j]-cExpected[j]
			if dc < -EPS || dc > EPS {
				t.Errorf("%d-order Legendre Polynomial incorrect, expecting %v, got %v",
					n, cExpected, cCalculated)
			}
		}
	}
}

func TestLegendreFunctions(t *testing.T) {
	ns := []int{2, 3, 4, 3, 6, 5, 7}
	ms := []int{0, 1, 2, 3, 2, 4, 3}
	xs := []float64{-0.9, 0.9, 0.15, -0.45, 0.65, 0.85, 0.45}
	vs := []float64{0.715, 1.994196267, -6.176578125, 10.68285409, -5.414123408, 61.85527031, -126.2222359}

	for i, vExpected := range vs {
		vCalculated := polynomial.LegendreFunction(ns[i], ms[i], xs[i])
		dv := vCalculated - vExpected
		if dv < -EPS || dv > EPS {
			t.Errorf("Legendre Function P[%d,%d] incorrect, expecting %4.1f, got %4.1f",
				ns[i], ms[i], vExpected, vCalculated)
		}
	}
}
