package polynomial

import (
	"math"
	"sync"
)

type legendreFunctionIndex struct {
	n, m int
}

var legendreFunctionCache sync.Map // map[legendreFunctionIndex]Polynomial

// LegendrePolynomial returns a Polynomial object corresponding to
// the Legendre Polynomial of degree n.
func LegendrePolynomial(n int) (p Polynomial) {
	p.c = make([]float64, n+1)
	for m := 0; m <= n/2; m++ {
		p.c[n-2*m] = Pow(-1, m) / Pow(2, n)
		p.c[n-2*m] *= FactorialRatioFloat(2*n-2*m, n-m) / float64(Factorial(m)*Factorial(n-2*m))
	}
	return p
}

// LegendreFunction evaluates the Associated Legendre Function at the given value.
// Normalization is that given in WMM2015_Report.pdf equation 6.
//
// The derived polynomial for each (n, m) is memoized in a process-wide cache,
// safe for concurrent use.
func LegendreFunction(n, m int, x float64) float64 {
	key := legendreFunctionIndex{n, m}
	var p Polynomial
	if v, ok := legendreFunctionCache.Load(key); ok {
		p = v.(Polynomial)
	} else {
		p = LegendrePolynomial(n).Derivative(m)
		legendreFunctionCache.Store(key, p)
	}
	return math.Pow(1-x*x, float64(m)/2) * p.Evaluate(x)
}
