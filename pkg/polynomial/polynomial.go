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

// Derivative calculates the polynomial corresponding to the nth derivative of the input polynomial.
func (p Polynomial) Derivative(n int) (q Polynomial) {
	if n==1 {
		q.c = make([]float64, len(p.c)-1)

		for m := 1; m < len(p.c); m++ {
			q.c[m-1] = float64(m) * p.c[m]
		}

		return q
	}

	q = p
	for m:=0; m<n; m++ {
		q = q.Derivative(1)
	}
	return q
}
