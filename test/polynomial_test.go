package main

import (
	"testing"

	"github.com/westphae/geomag/pkg/polynomial"
)

const EPS = 1e-8

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


