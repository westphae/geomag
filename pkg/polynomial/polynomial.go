package polynomial

type Polynomial struct {
	c []float64
}

// NewPolynomial makes a new polynomial object with the specified coefficients.
// e.g. for x^2-1, use NewPolynomial([]float64{-1,0,1}.
func NewPolynomial(c []float64) (p Polynomial) {
	p.c = c
	return p
}

// Coefficients returns the coefficients of the polynomial in a slice.
func (p Polynomial) Coefficients() (c []float64) {
	return p.c
}

// Evaluate calculates the value of the polynomial at the given input value.
func (p Polynomial) Evaluate(x float64) (y float64) {
	for m, c := range p.c {
		y += c*Pow(x, m)
	}

	return y
}

// Derivative calculates the polynomial corresponding to the nth derivative
// of the input polynomial. n must be non-negative; n=0 returns p unchanged.
// For negative n, returns p unchanged (rather than recursing forever or
// producing a zero polynomial of nonsensical degree).
func (p Polynomial) Derivative(n int) (q Polynomial) {
	if n <= 0 {
		return p
	}
	if n == 1 {
		// Differentiating a constant (or empty polynomial) yields 0.
		if len(p.c) <= 1 {
			return NewPolynomial([]float64{0})
		}
		q.c = make([]float64, len(p.c)-1)
		for m := 1; m < len(p.c); m++ {
			q.c[m-1] = float64(m) * p.c[m]
		}
		return q
	}

	q = p
	for range n {
		q = q.Derivative(1)
	}
	return q
}
