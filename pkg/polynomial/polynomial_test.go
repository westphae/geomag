package polynomial

import (
	"fmt"
	"testing"
)

const eps = 1e-6

func testDiff(name string, actual, expected float64, eps float64, t *testing.T) {
	if actual - expected > -eps && actual - expected < eps {
		t.Logf("%s correct: expected %8.4f, got %8.4f", name, expected, actual)
		return
	}
	t.Errorf("%s incorrect: expected %8.4f, got %8.4f", name, expected, actual)
}

func TestPow(t *testing.T) {
	var (
		xs = []float64{2.0, 0.5, 1.0, 3.14, 10}
		ns = []int{5, 3, 4, 0, -3}
		ys = []float64{32, 0.125, 1, 1, 0.001}
	)

	for i:=0; i<len(xs); i++ {
		y := Pow(xs[i], ns[i])
		testDiff("Pow", y, ys[i], eps, t)
	}
}

func TestFactorial(t *testing.T) {
	var (
		ns = []int{20, 19, 5, 3, 4, 0, 1}
		zs = []int{2432902008176640000, 121645100408832000, 120, 6, 24, 1, 1}
	)

	for i:=0; i<len(ns); i++ {
		z := Factorial(ns[i])
		testDiff(fmt.Sprintf("%d!", ns[i]), float64(z), float64(zs[i]), eps, t)
	}
}

// FactorialRatioFloat needs to calculate up to 24!
func TestFactorialRatioFloat(t *testing.T) {
	var (
		ns = []int{6, 6, 6, 6, 3, 3, 1, 24}
		ms = []int{2, 3, 1, 0, 3, 2, 1, 0}
		zs = []float64{360, 120, 720, 720, 1, 3, 1, 620448401733239439360000}
	)

	for i:=0; i<len(ns); i++ {
		z := FactorialRatioFloat(ns[i], ms[i])
		testDiff(fmt.Sprintf("%d!/%d!", ns[i], ms[i]), z, zs[i], eps, t)
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
			p := NewPolynomial(cs[i])
			y := p.Evaluate(xs[j])
			testDiff(fmt.Sprintf("Evaluate %v(%3.1f)", cs[i], xs[j]), y, ys[j][i], eps, t)
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
		p := NewPolynomial(cs[i])

		y := p.Derivative(1).Coefficients()
		for j, d := range ds[i] {
			testDiff(fmt.Sprintf("Derivative of %v", cs[i]), y[j], d, eps, t)
		}

		y = p.Derivative(2).Coefficients()
		for j, d := range dds[i] {
			testDiff(fmt.Sprintf("Second derivative of %v", cs[i]), y[j], d, eps, t)
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
		cCalculated := LegendrePolynomial(n).Coefficients()
		for j:=0; j<=n; j++ {
			testDiff(fmt.Sprintf("Order-%d Legendre Polynomial", n), cCalculated[j], cExpected[j], eps, t)
		}
	}
}

func TestLegendreFunctions(t *testing.T) {
	ns := []int{2, 3, 4, 3, 6, 5, 7}
	ms := []int{0, 1, 2, 3, 2, 4, 3}
	xs := []float64{-0.9, 0.9, 0.15, -0.45, 0.65, 0.85, 0.45}
	vs := []float64{0.715, 1.994196267, -6.176578125, 10.68285409, -5.414123408, 61.85527031, -126.2222359}

	for i, vExpected := range vs {
		vCalculated := LegendreFunction(ns[i], ms[i], xs[i])
		testDiff(fmt.Sprintf("Legendre function P(%d,%d)", ns[i], ms[i]), vCalculated, vExpected, eps, t)
	}
}
