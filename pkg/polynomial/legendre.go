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
//
// Coefficients are computed in float64 throughout: the constant-factor
// products `m! * (n-2m)!` exceed int64's range for n>~20 (Factorial(21)
// already overflows), so we use FactorialFloat / FactorialRatioFloat
// to keep things finite up to roughly n=170. Beyond that the float64
// exponent runs out of headroom and the polynomial coefficients begin
// to lose precision; for the WMM family (n≤12) and WMMHR (n=133) we're
// safely within range.
func LegendrePolynomial(n int) (p Polynomial) {
	p.c = make([]float64, n+1)
	twoN := Pow(2, n)
	for m := 0; m <= n/2; m++ {
		c := Pow(-1, m) / twoN
		c *= FactorialRatioFloat(2*n-2*m, n-m) / (FactorialFloat(m) * FactorialFloat(n-2*m))
		p.c[n-2*m] = c
	}
	return p
}

// LegendreFunction evaluates the Associated Legendre Function at the given value.
// Normalization is that given in WMM2015_Report.pdf equation 6.
//
// The derived polynomial for each (n, m) is memoized in a process-wide cache,
// safe for concurrent use.
//
// Numerically stable for n up to roughly 20. For higher n (e.g. WMMHR at
// n=133) the underlying polynomial-coefficient computation suffers
// catastrophic cancellation; use SchmidtNormalizedALFTable instead.
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

// SchmidtNormalizedALFTable evaluates all Schmidt semi-normalized associated
// Legendre functions P̄_n^m(x) and their derivatives with respect to
// latitude dP̄_n^m/dφ for 0 ≤ m ≤ n ≤ nMax, at x = sin(φ).
//
// Returns two triangular tables P and dP where P[n][m] (and dP[n][m]) is
// valid for 0 ≤ m ≤ n ≤ nMax; access outside that range is undefined.
//
// Implementation is a Go port of NOAA's MAG_PcupHigh (GeomagnetismLibrary.c
// in the WMM reference distribution), which itself follows Holmes &
// Featherstone, "A unified approach to the Clenshaw summation and the
// recursive computation of very high degree and order normalised
// associated Legendre functions" (J. Geodesy 76, 2002, 279-299). The
// recurrence is numerically stable for n up to thousands; for the WMM
// family at n=12 and WMMHR at n=133 it is well within range.
//
// The 1e-280 scale factor on the diagonal P̄_m^m terms (rescaled out at
// the end of each m-row) prevents underflow at large m near the poles
// where (1-x²)^(m/2) would otherwise vanish.
//
// Returns nil tables if |x| == 1 (the derivatives are undefined at the
// geographic poles in this representation).
func SchmidtNormalizedALFTable(x float64, nMax int) (P, dP [][]float64) {
	if nMax < 0 {
		return nil, nil
	}
	if math.Abs(x) >= 1 {
		// dP/dφ has a 1/cos(φ) factor that diverges at the poles.
		return nil, nil
	}
	z := math.Sqrt((1.0 - x) * (1.0 + x)) // = cos(latitude)

	// Pre-square-root cache: PreSqr[i] = sqrt(i) for i in [0, 2*nMax+1].
	preSqr := make([]float64, 2*nMax+2)
	for i := range preSqr {
		preSqr[i] = math.Sqrt(float64(i))
	}

	// Allocate triangular tables P[n][m] and dP[n][m] for 0 ≤ m ≤ n ≤ nMax.
	P = make([][]float64, nMax+1)
	dP = make([][]float64, nMax+1)
	for n := 0; n <= nMax; n++ {
		P[n] = make([]float64, n+1)
		dP[n] = make([]float64, n+1)
	}

	// f1, f2: per-(n,m) recurrence coefficients precomputed. We store
	// them flattened on the same triangular layout for tightness.
	f1 := make([][]float64, nMax+1)
	f2 := make([][]float64, nMax+1)
	for n := 2; n <= nMax; n++ {
		f1[n] = make([]float64, n+1)
		f2[n] = make([]float64, n+1)
		// m = 0 case
		f1[n][0] = float64(2*n-1) / float64(n)
		f2[n][0] = float64(n-1) / float64(n)
		// m in [1, n-2]
		for m := 1; m <= n-2; m++ {
			f1[n][m] = float64(2*n-1) / preSqr[n+m] / preSqr[n-m]
			f2[n][m] = preSqr[n-m-1] * preSqr[n+m-1] / preSqr[n+m] / preSqr[n-m]
		}
	}

	// Initial column m=0: standard Legendre recurrence P_n^0.
	P[0][0] = 1.0
	dP[0][0] = 0.0
	if nMax == 0 {
		return P, dP
	}
	pm1 := x
	pm2 := 1.0
	P[1][0] = pm1
	dP[1][0] = z

	for n := 2; n <= nMax; n++ {
		plm := f1[n][0]*x*pm1 - f2[n][0]*pm2
		P[n][0] = plm
		dP[n][0] = float64(n) * (pm1 - x*plm) / z
		pm2 = pm1
		pm1 = plm
	}

	// Off-diagonal columns m ≥ 1: walk down the diagonal P_m^m, then
	// step to P_{m+1}^m, then recurrence to P_n^m for n > m+1.
	const scalef = 1.0e-280
	pmm := preSqr[2] * scalef // unscaled P̄_1^1 = sqrt(2) * z (z absorbed into rescalem below)
	rescalem := 1.0 / scalef

	for m := 1; m <= nMax-1; m++ {
		rescalem *= z

		// P̄_m^m = sqrt((2m+1)/(2m)) * z * P̄_{m-1}^{m-1}
		// (with the diagonal scale factor still applied)
		pmm *= preSqr[2*m+1] / preSqr[2*m]
		P[m][m] = pmm * rescalem / preSqr[2*m+1]
		dP[m][m] = -float64(m) * x * P[m][m] / z
		pm2 = pmm / preSqr[2*m+1]

		// P̄_{m+1}^m = sqrt(2m+1) * x * P̄_m^m
		pm1 = x * preSqr[2*m+1] * pm2
		P[m+1][m] = pm1 * rescalem
		dP[m+1][m] = ((pm2 * rescalem) * preSqr[2*m+1] - x * float64(m+1) * P[m+1][m]) / z

		// P̄_n^m for n > m+1 via the standard 2-term recurrence
		for n := m + 2; n <= nMax; n++ {
			plm := x*f1[n][m]*pm1 - f2[n][m]*pm2
			P[n][m] = plm * rescalem
			dP[n][m] = (preSqr[n+m]*preSqr[n-m]*(pm1*rescalem) - float64(n)*x*P[n][m]) / z
			pm2 = pm1
			pm1 = plm
		}
	}

	// Final diagonal P̄_{nMax}^{nMax}.
	if nMax >= 1 {
		rescalem *= z
		pmm /= preSqr[2*nMax]
		P[nMax][nMax] = pmm * rescalem
		dP[nMax][nMax] = -float64(nMax) * x * P[nMax][nMax] / z
	}

	return P, dP
}
